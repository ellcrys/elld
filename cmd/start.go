package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/params"
	"github.com/pkg/profile"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/miner"
	"github.com/ellcrys/elld/rpc"

	"gopkg.in/asaskevich/govalidator.v4"

	"github.com/ellcrys/elld/accountmgr"
	funk "github.com/thoas/go-funk"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/console"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/util"
	"github.com/spf13/cobra"
)

// getKey unlocks an account and returns the corresponding key.
func getKey(accountID, password string, seed int64) (*crypto.Key, error) {

	var key *crypto.Key
	var err error
	var storedAccount *accountmgr.StoredAccount

	// Check whether the account id is actually a
	// private key, if it is, we return a key.
	if err := crypto.IsValidPrivKey(accountID); err == nil {
		key, _ := crypto.PrivKeyFromBase58(accountID)
		return crypto.NewKeyFromPrivKey(key), nil
	}

	// If an account id is provided...
	if accountID != "" {
		// and it is numeric we assume the caller is referring
		// to the index position their account is occupying on disk.
		if govalidator.IsNumeric(accountID) {
			i, err := strconv.Atoi(accountID)
			if err != nil {
				return nil, err
			}
			// Get the account by index
			storedAccount, err = accountMgr.GetByIndex(i)
			if err != nil {
				return nil, err
			}
		} else {
			// Here we assume the user provided an address, so
			// we fetch the account by the address
			storedAccount, err = accountMgr.GetByAddress(accountID)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// At this point, the user did not specify an identifier
		// for an account they want, so we will create a random
		// address and tag it as an ephemeral address
		key, _ = accountMgr.CreateCmd(0, util.RandString(32))
		key.Meta["ephemeral"] = true
		return key, nil
	}

	// If the password is unset, start an interactive session to
	// request password from user.
	if password == "" {
		fmt.Println(fmt.Sprintf("Account {%s} needs to be unlocked. Please enter your password.", storedAccount.Address))
		password, err = accountMgr.AskForPasswordOnce()
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	// If password is set and it's a path to a file,
	// read the file and use its content as the password
	if len(password) > 1 && (os.IsPathSeparator(password[0]) || (len(password) >= 2 && password[:2] == "./")) {
		content, err := ioutil.ReadFile(password)
		if err != nil {
			if funk.Contains(err.Error(), "no such file") {
				return nil, fmt.Errorf("password file {%s} not found", password)
			}
			if funk.Contains(err.Error(), "is a directory") {
				return nil, fmt.Errorf("password file path {%s} is a directory. Expects a file", password)
			}
			return nil, err
		}
		password = string(content)
		password = strings.TrimSpace(strings.Trim(password, "/n"))
	}

	// Use the password to decrypt the account
	if err = storedAccount.Decrypt(password); err != nil {
		return nil, fmt.Errorf("account unlock failed. %s", err)
	}

	return storedAccount.GetAddress(), nil
}

// starts the node.
// - Parse flags
// - Validate node bind address
// - Load an account
// - create local node
// - add hardcoded node as bootstrap node if any
// - add bootstrap node from config file if any
// - open database
// - initialize protocol instance along with message handlers
// - create global event handler
// - create logic handler and pass it to the node
// - start RPC server if enabled
// - start console if enabled
// - connect console to rpc server and prepare console vm if rpc server is enabled
func start(cmd *cobra.Command, args []string, startConsole bool) (*node.Node, *rpc.Server, *console.Console, *miner.Miner) {

	var err error

	// Process flags
	viper.BindPFlag("node.account", cmd.Flags().Lookup("account"))
	viper.BindPFlag("node.password", cmd.Flags().Lookup("pwd"))
	viper.BindPFlag("node.bootstrapAddrs", cmd.Flags().Lookup("add-node"))
	viper.BindPFlag("node.address", cmd.Flags().Lookup("address"))
	viper.BindPFlag("rpc.enabled", cmd.Flags().Lookup("rpc"))
	viper.BindPFlag("rpc.address", cmd.Flags().Lookup("rpc-address"))
	viper.BindPFlag("rpc.disableAuth", cmd.Flags().Lookup("rpc-disable-auth"))
	viper.BindPFlag("node.seed", cmd.Flags().Lookup("seed"))
	viper.BindPFlag("miner.enabled", cmd.Flags().Lookup("mine"))
	viper.BindPFlag("miner.numMiners", cmd.Flags().Lookup("miners"))
	viper.BindPFlag("node.noNet", cmd.Flags().Lookup("no-net"))
	account := viper.GetString("node.account")
	password := viper.GetString("node.password")
	listeningAddr := viper.GetString("node.address")
	startRPC := viper.GetBool("rpc.enabled")
	rpcAddress := viper.GetString("rpc.address")
	seed := viper.GetInt64("node.seed")
	mine := viper.GetBool("miner.enabled")
	numMiners := viper.GetInt("miner.numMiners")
	noNet := viper.GetBool("node.noNet")

	// Unmarshal configurations known to viper into our
	// config object.
	if err := viper.Unmarshal(&(*cfg)); err != nil {
		log.Fatal("Failed to unmarshal configuration file: %s", err)
	}

	// check that the host address to bind
	// the engine to is valid,
	if !util.IsValidHostPortAddress(listeningAddr) {
		log.Fatal("invalid bind address provided")
	}

	// Load the coinbase account.
	// Required for signing blocks, transactions and
	// for receiving mining rewards.
	coinbase, err := getKey(account, password, seed)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Prevent mining when the node's key is ephemeral
	if mine && coinbase.Meta["ephemeral"] != nil {
		log.Fatal(params.ErrMiningWithEphemeralKey.Error())
	}

	// Create event the global event handler
	event := &emitter.Emitter{}

	// Create the local node.
	n, err := node.NewNode(cfg, listeningAddr, coinbase, log)
	if err != nil {
		log.Fatal("failed to create local node", "Err", err.Error())
	}

	log.Info("Elld has started",
		"ClientVersion", cfg.VersionInfo.BuildVersion,
		"NetVersion", config.Versions.Protocol,
		"DevMode", devMode,
		"Name", n.Name)

	if noNet {
		n.GetHost().Close()
		n.NoNetwork()
	}

	// In debug mode, we set log level to DEBUG.
	if n.DevMode() {
		log.SetToDebug()
	}

	// Configure transactions pool and assign to node
	pool := txpool.New(params.PoolCapacity)
	n.SetTxsPool(pool)

	if !noNet {
		// Add hardcoded bootstrap addresses
		if err := n.AddAddresses(config.SeedAddresses, true); err != nil {
			log.Fatal("%s", err)
		}

		// Add bootstrap addresses supplied in the config file
		if err := n.AddAddresses(cfg.Node.BootstrapAddresses, false); err != nil {
			log.Fatal("%s", err)
		}
	}

	// open the database on the engine
	if err = n.OpenDB(); err != nil {
		log.Fatal("failed to open local database", "Err", err.Error())
	}

	log.Info("Ready for connections", "Addr", n.GetAddress().ConnectionString())

	// Initialize and set the blockchain manager's db,
	// event emitter and pass it to the engine
	bChain := blockchain.New(n.GetTxPool(), cfg, log)
	bChain.SetDB(n.DB())
	bChain.SetEventEmitter(event)
	bChain.SetCoinbase(coinbase)

	// Initialize the miner, rpc server
	miner := miner.NewMiner(coinbase, bChain, event, cfg, log)
	rpcServer := rpc.NewServer(n.DB(), rpcAddress, cfg, log)

	// Set the node's references
	n.SetBlockchain(bChain)
	n.SetEventEmitter(event)

	// Setup block manager
	bm := node.NewBlockManager(n)
	bm.SetMiner(miner)
	go bm.Manage()
	n.SetBlockManager(bm)

	// Set transaction manager
	tm := node.NewTxManager(n)
	go tm.Manage()
	n.SetTxManager(tm)

	// power up the blockchain manager
	if err := bChain.Up(); err != nil {
		log.Fatal("failed to load blockchain manager", "Err", err.Error())
	}

	// Start the block manager and the node
	n.Start()

	// Initialized and start the miner if enabled via the cli flag.
	miner.SetNumThreads(numMiners)
	if mine {
		go miner.Begin()
	}

	// Add the RPC APIs from various components.
	rpcServer.AddAPI(
		n.APIs(),
		miner.APIs(),
		accountMgr.APIs(),
		bChain.APIs(),
		rpcServer.APIs(),
	)

	if startRPC {
		go rpcServer.Serve()
	}

	// Initialize and start the console if
	// enabled via the appropriate cli flag.
	var cs *console.Console
	if startConsole {

		// Create the console.
		// Configure the RPC client if the server has started
		cs = console.New(coinbase, consoleHistoryFilePath, cfg, log)
		cs.SetVersions(config.Versions.Protocol, BuildVersion, GoVersion, BuildCommit)
		cs.SetRPCServer(rpcServer, false)

		// Prepare the console
		if err := cs.Prepare(); err != nil {
			log.Fatal("failed to prepare console VM", "Err", err)
		}

		// Run the console.
		fmt.Println("") // Extra space in console
		go cs.Run()
	}

	return n, rpcServer, cs, miner
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the node",
	Long: `Description:
  Starts a node.
  
  Set the listening address on the node using '--address' flag. 
  
  Use '--addnode' to provide a comma separated list of initial addresses of peers
  to connect to. Addresses must be valid ipfs multiaddress. An account must be 
  provided and unlocked to be used for signing transactions and blocks. Use '--account'
  flag to provide the account. If account is not provided, the default account (
  oldest account) in <CONFIGDIR>/` + config.AccountDirName + ` is used instead.
  
  If no account was found, an interactive session to create an account is started.   
  
  Account password will be interactively requested during account creation and unlock
  operations. Use '--pwd' flag to provide the account password non-interactively. '--pwd'
  can also accept a path to a file containing the password.`,
	Run: func(cmd *cobra.Command, args []string) {

		profilePath := profile.ProfilePath(cfg.NetDataDir())
		viper.BindPFlag("debug.cpuProfile", cmd.Flags().Lookup("cpu-profile"))
		viper.BindPFlag("debug.memProfile", cmd.Flags().Lookup("mem-profile"))
		viper.BindPFlag("debug.mutexProfile", cmd.Flags().Lookup("mutex-profile"))
		cpuProfile := viper.GetBool("debug.cpuProfile")
		memProfile := viper.GetBool("debug.memProfile")
		mtxProfile := viper.GetBool("debug.mutexProfile")

		if cpuProfile {
			defer profile.Start(profile.CPUProfile, profilePath).Stop()
		}

		if memProfile {
			defer profile.Start(profile.MemProfile, profilePath).Stop()
		}

		if mtxProfile {
			defer profile.Start(profile.MutexProfile, profilePath).Stop()
		}

		n, rpcServer, _, miner := start(cmd, args, false)

		setTerminateFunc(func() {
			if miner != nil {
				miner.Stop()
			}
			if rpcServer != nil {
				rpcServer.Stop()
			}
			if n != nil {
				n.Stop()
			}
		})

		n.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringSliceP("add-node", "j", nil, "Add the address of a node to connect to.")
	startCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address local node will listen on.")
	startCmd.Flags().Bool("rpc", false, "Enables the RPC server")
	startCmd.Flags().String("rpc-address", "127.0.0.1:8999", "Address RPC server will listen on.")
	startCmd.Flags().Bool("rpc-disable-auth", false, "Disable RPC authentication (not recommended)")
	startCmd.Flags().String("account", "", "Coinbase account to load. An ephemeral account is used as default.")
	startCmd.Flags().String("pwd", "", "The password of the node's network account.")
	startCmd.Flags().Int64P("seed", "s", 0, "Provide a strong seed for network account creation (not recommended)")
	startCmd.Flags().Bool("mine", false, "Start Blake2 CPU mining")
	startCmd.Flags().Int("miners", 0, "The number of miner threads to use. (Default: Number of CPU)")
	startCmd.Flags().Bool("no-net", false, "Closes the network host and prevents (in/out) connections")
}

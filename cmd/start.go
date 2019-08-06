package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ellcrys/elld/elldb"

	"github.com/ellcrys/elld/burner"

	"github.com/spf13/viper"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/ltcsuite/ltcd"
	"github.com/ellcrys/elld/params"
	"github.com/pkg/profile"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain"
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
	if err = storedAccount.Decrypt(password, false); err != nil {
		return nil, fmt.Errorf("account unlock failed. %s", err)
	}

	return storedAccount.GetKey().(*crypto.Key), nil
}

// makeBurnerChainArgs returns arguments
// compatible with the burn chain
func makeBurnerChainArgs() []string {
	args := []string{}
	if viper.GetBool("burner.testnet") {
		args = append(args, "--testnet")
	}
	if viper.GetBool("burner.notls") {
		args = append(args, "--notls")
	}
	if rpcUser := viper.GetString("burner.rpcuser"); len(rpcUser) > 0 {
		args = append(args, []string{"--rpcuser", rpcUser}...)
	}
	if rpcPass := viper.GetString("burner.rpcpass"); len(rpcPass) > 0 {
		args = append(args, []string{"--rpcpass", rpcPass}...)
	}
	if rpcListen := viper.GetString("burner.rpclisten"); len(rpcListen) > 0 {
		args = append(args, []string{"--rpclisten", rpcListen}...)
	}
	return args
}

// initBurnerServer sets up the burner server.
// It will return nil if the server stopped gracefully
// or error if it was interrupted.
func initBurnerServer(cmd *cobra.Command, db elldb.DB) chan error {

	viper.BindPFlag("burner.testnet", cmd.Flags().Lookup("burner-testnet"))
	viper.BindPFlag("burner.rpcuser", cmd.Flags().Lookup("burner-rpcuser"))
	viper.BindPFlag("burner.rpcpass", cmd.Flags().Lookup("burner-rpcpass"))
	viper.BindPFlag("burner.notls", cmd.Flags().Lookup("burner-notls"))
	viper.BindPFlag("burner.rpclisten", cmd.Flags().Lookup("burner-rpclisten"))
	viper.BindPFlag("burner.utxokeeperskip", cmd.Flags().Lookup("burner-utxokeeperskip"))
	viper.BindPFlag("burner.utxokeeperworkers", cmd.Flags().Lookup("burner-utxokeeperworkers"))
	viper.BindPFlag("burner.utxokeeperoff", cmd.Flags().Lookup("burner-utxokeeperoff"))
	viper.BindPFlag("burner.utxokeeperreindex", cmd.Flags().Lookup("burner-utxokeeperreindex"))

	testnet := viper.GetBool("burner.testnet")
	noTLS := viper.GetBool("burner.notls")
	rpcUser := viper.GetString("burner.rpcuser")
	rpcPass := viper.GetString("burner.rpcpass")
	rpcListen := viper.GetString("burner.rpclisten")
	utxoKeeperSkip := viper.GetInt32("burner.utxokeeperskip")
	utxoKeeperNumThread := viper.GetInt("burner.utxokeeperworkers")
	noUTXOKeeper := viper.GetBool("burner.utxokeeperoff")
	reIndex := viper.GetBool("burner.utxokeeperreindex")

	// Set default burner RPC listening address
	if len(rpcListen) == 0 {
		rpcListen = "127.0.0.1:9334"
		if testnet {
			rpcListen = "127.0.0.1:19334"
		}
	}

	// Configure burn chain argument and start the burn chain server.
	// The status channel is used inform other processes about errors
	// that caused the burner server to stop.
	os.Args = append([]string{""}, makeBurnerChainArgs()...)
	config.SetBurnerMainnet(!testnet)
	status := make(chan error)
	go ltcd.Main(interrupt, status)

	// Ensure the burner server did not fail to start.
	// If it did, terminate the program using os.Exit(1)
	err := <-status
	if err != nil {
		os.Exit(1)
	}

	// If the burner RPC user and pass are provided, then
	// we need to start the burner account utxo indexer
	if rpcUser != "" && rpcPass != "" {
		if !noUTXOKeeper {
			bc, err := burner.GetClient(rpcListen, rpcUser, rpcPass, noTLS)
			if err != nil {
				close(interrupt)
				log.Fatal("failed to setup burner RPC server client", "Err", err.Error())
			}

			utxoKeeper := burner.NewBurnerAccountUTXOKeeper(log, db, config.GetNetVersion(), interrupt)
			utxoKeeper.SetClient(bc)
			go utxoKeeper.Begin(accountMgr, utxoKeeperNumThread, utxoKeeperSkip, reIndex)
		}
	}

	return status
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
func start(cmd *cobra.Command, args []string, startConsole bool, interrupt chan struct{}) {

	var err error

	// Process flags
	viper.BindPFlag("node.account", cmd.Flags().Lookup("account"))
	viper.BindPFlag("node.password", cmd.Flags().Lookup("pwd"))
	viper.BindPFlag("node.bootstrapAddrs", cmd.Flags().Lookup("add-node"))
	viper.BindPFlag("node.address", cmd.Flags().Lookup("address"))
	viper.BindPFlag("rpc.enabled", cmd.Flags().Lookup("rpc"))
	viper.BindPFlag("rpc.address", cmd.Flags().Lookup("rpc-address"))
	viper.BindPFlag("rpc.disableAuth", cmd.Flags().Lookup("rpc-disable-auth"))
	viper.BindPFlag("rpc.sessionTTL", cmd.Flags().Lookup("rpc-session-ttl"))
	viper.BindPFlag("node.seed", cmd.Flags().Lookup("seed"))
	viper.BindPFlag("node.noNet", cmd.Flags().Lookup("no-net"))
	viper.BindPFlag("node.syncDisabled", cmd.Flags().Lookup("sync-disabled"))

	// Host chain config variables
	account := viper.GetString("node.account")
	password := viper.GetString("node.password")
	listeningAddr := viper.GetString("node.address")
	startRPC := viper.GetBool("rpc.enabled")
	rpcAddress := viper.GetString("rpc.address")
	seed := viper.GetInt64("node.seed")
	noNet := viper.GetBool("node.noNet")
	syncDisabled := viper.GetBool("node.syncDisabled")

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

	// Create event the global event handler
	event := &emitter.Emitter{}

	// Create the local node.
	n, err := node.NewNode(cfg, listeningAddr, coinbase, log)
	if err != nil {
		log.Fatal("failed to create local node", "Err", err.Error())
	}

	// In debug mode, we set log level to DEBUG.
	if n.DevMode() {
		log.SetToDebug()
	}

	// Set sync mode
	n.SetSyncMode(node.NewDefaultSyncMode(syncDisabled))

	// Disable network if required
	if noNet {
		n.GetHost().Close()
		n.DisableNetwork()
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

	log.Info("Elld has started",
		"ClientVersion", cfg.VersionInfo.BuildVersion,
		"NetVersion", config.GetVersions().Protocol,
		"DevMode", devMode,
		"SyncEnabled", !n.GetSyncMode().IsDisabled(),
		"NetworkEnabled", !noNet,
		"Name", n.Name)

	log.Info("Ready for connections", "Addr", n.GetAddress().ConnectionString())

	// Initialize and set the blockchain manager's db,
	// event emitter and pass it to the engine
	bChain := blockchain.New(n.GetTxPool(), cfg, log)
	bChain.SetDB(n.DB())
	bChain.SetInterrupt(interrupt)
	bChain.SetEventEmitter(event)
	bChain.SetCoinbase(coinbase)

	// Initialize rpc server
	rpcServer := rpc.NewServer(n.DB(), rpcAddress, cfg, log, interrupt)

	// Set the node's references
	n.SetBlockchain(bChain)
	n.SetEventEmitter(event)
	n.SetInterrupt(interrupt)

	// Setup block manager
	bm := node.NewBlockManager(n)
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

	// Initialize the burner chain
	burnerStatus := initBurnerServer(cmd, n.DB())

	// Start the block manager and the node
	n.Start()

	// Create a burner API
	burnerAPI := burner.NewAPI()

	// Add the RPC APIs from various components.
	rpcServer.AddAPI(
		n.APIs(),
		accountMgr.APIs(),
		bChain.APIs(),
		rpcServer.APIs(),
		burnerAPI.APIs(),
	)

	// Start the JSON-RPC 2.0 server and wait for it to start
	if startRPC {
		go rpcServer.Serve()
		time.Sleep(1 * time.Second)
	}

	// Initialize and start the console if
	// enabled via the appropriate cli flag.
	var cs *console.Console
	if startConsole {

		// Create the console.
		// Configure the RPC client if the server has started
		cs = console.New(coinbase, consoleHistoryFilePath, cfg, log)
		cs.SetVersions(config.GetVersions().Protocol, BuildVersion, GoVersion, BuildCommit)
		cs.SetRPCServer(rpcServer, false)

		// Prepare the console
		if err := cs.Prepare(); err != nil {
			log.Fatal("failed to prepare console VM", "Err", err)
		}

		// Run the console.
		go cs.Run()

		cs.OnStop(func() {
			if !util.IsStructChanClosed(interrupt) {
				close(interrupt)
			}
		})
	}

	<-burnerStatus
	n.Wait()

	return
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

		// Start the local node and other subservices
		start(cmd, args, false, interrupt)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringSliceP("add-node", "j", nil, "Add the address of a node to connect to.")
	startCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address local node will listen on.")
	startCmd.Flags().Bool("rpc", false, "Enables the RPC server")
	startCmd.Flags().String("rpc-address", "127.0.0.1:8999", "Address RPC server will listen on.")
	startCmd.Flags().Bool("rpc-disable-auth", false, "Disable RPC authentication (not recommended)")
	startCmd.Flags().Int64("rpc-session-ttl", 3600000, "The time-to-live (in milliseconds) of RPC session tokens")
	startCmd.Flags().String("account", "", "Coinbase account to load. An ephemeral account is used as default.")
	startCmd.Flags().String("pwd", "", "The password of the node's network account.")
	startCmd.Flags().Int64P("seed", "s", 0, "Provide a strong seed for network account creation (not recommended)")
	startCmd.Flags().Bool("no-net", false, "Closes the network host and prevents (in/out) connections")
	startCmd.Flags().Bool("sync-disabled", false, "Disable block and transaction synchronization")

	// Burner chain related flags
	startCmd.Flags().Bool("burner-testnet", false, "Run the burner server on the testnet")
	startCmd.Flags().String("burner-rpcuser", "", "RPC username of the burner server")
	startCmd.Flags().String("burner-rpcpass", "", "RPC password of the burner server")
	startCmd.Flags().Bool("burner-notls", false, "Run the burner server on the testnet")
	startCmd.Flags().String("burner-rpclisten", "", "Set the burner RPC server interface/port to listen for connections.")
	startCmd.Flags().Int32("burner-utxokeeperskip", 0, "Force the burner account utxo keeper to skip blocks below the given height.")
	startCmd.Flags().Int("burner-utxokeeperworkers", 3, "Set the number of burner account UTXO keeper worker threads.")
	startCmd.Flags().Bool("burner-utxokeeperoff", false, "Disable the burner account UTXO keeper service.")
	startCmd.Flags().Bool("burner-utxokeeperreindex", false, "Force the UTXO keeper to re-index burner accounts")
}

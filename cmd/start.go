package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

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

var (
	boostrapAddresses = []string{}
)

func devDefaultConfig(cfg *config.EngineConfig) {
	cfg.Node.GetAddrInterval = util.NonZeroOrDefIn64(cfg.Node.GetAddrInterval, 10)
	cfg.Node.PingInterval = util.NonZeroOrDefIn64(cfg.Node.PingInterval, 60)
	cfg.Node.SelfAdvInterval = util.NonZeroOrDefIn64(cfg.Node.SelfAdvInterval, 10)
	cfg.Node.CleanUpInterval = util.NonZeroOrDefIn64(cfg.Node.CleanUpInterval, 10)
	cfg.Node.ConnEstInterval = util.NonZeroOrDefIn64(cfg.Node.ConnEstInterval, 10)
	cfg.TxPool.Capacity = util.NonZeroOrDefIn64(cfg.TxPool.Capacity, 100)
}

// loadOrCreateAccount unlocks an account and returns the underlying address.
// - If account is provided, it is fetched and unlocked using the password provided.
//	 If password is not provided, the is requested through an interactive prompt.
// - If account is not provided, an ephemeral key is created and returned.
func loadOrCreateAccount(accountID, password string, seed int64) (*crypto.Key, error) {

	var address *crypto.Key
	var err error
	var storedAccount *accountmgr.StoredAccount

	if accountID != "" {
		if govalidator.IsNumeric(accountID) {
			aInt, err := strconv.Atoi(accountID)
			if err != nil {
				return nil, err
			}
			storedAccount, err = accountMgr.GetByIndex(aInt)
			if err != nil {
				return nil, err
			}
		} else {
			storedAccount, err = accountMgr.GetByAddress(accountID)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// create ephemeral account
		address, _ = accountMgr.CreateCmd(0, util.RandString(32))
		address.Meta["ephemeral"] = true
		return address, nil
	}

	// If the password is unset, request password from user
	if password == "" {
		fmt.Println(fmt.Sprintf("Account {%s} needs to be unlocked. Please enter your password.", storedAccount.Address))
		password, err = accountMgr.AskForPasswordOnce()
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	// If password is set and its a path to a file,
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

	if err = storedAccount.Decrypt(password); err != nil {
		return nil, fmt.Errorf("account unlock failed. %s", err)
	}

	return storedAccount.GetAddress(), nil
}

// starts the node.
// - Parse flags
// - Set default configurations
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
	bootstrapAddresses, _ := cmd.Flags().GetStringSlice("addnode")
	listeningAddr, _ := cmd.Flags().GetString("address")
	startRPC, _ := cmd.Flags().GetBool("rpc")
	rpcAddress, _ := cmd.Flags().GetString("rpcaddress")
	account, _ := cmd.Flags().GetString("account")
	password, _ := cmd.Flags().GetString("pwd")
	seed, _ := cmd.Flags().GetInt64("seed")
	mine, _ := cmd.Flags().GetBool("mine")
	numMiners, _ := cmd.Flags().GetInt("miners")
	debug, _ := cmd.Flags().GetBool("debug")

	if len(account) == 0 {
		account = os.Getenv("ELLD_ACCOUNT")
	}

	if len(password) == 0 {
		password = os.Getenv("ELLD_ACCOUNT_PASSWORD")
	}

	if os.Getenv("ELLD_RPC_ON") == "true" {
		startRPC = true
	}

	if addr := os.Getenv("ELLD_RPC_ADDRESS"); len(addr) > 0 {
		rpcAddress = addr
	}

	if addr := os.Getenv("ELLD_LADDRESS"); len(addr) > 0 {
		listeningAddr = addr
	}

	if envAddNode := os.Getenv("ELLD_ADDNODE"); len(envAddNode) > 0 {
		addrs := strings.Split(envAddNode, ",")
		cfg.Node.BootstrapAddresses = append(cfg.Node.BootstrapAddresses, addrs...)
	}

	// Set configurations
	cfg.Node.MessageTimeout = util.NonZeroOrDefIn64(cfg.Node.MessageTimeout, 30)
	cfg.Node.BootstrapAddresses = append(cfg.Node.BootstrapAddresses, bootstrapAddresses...)
	cfg.Node.MaxAddrsExpected = 1000
	cfg.Node.MaxOutboundConnections = util.NonZeroOrDefIn64(cfg.Node.MaxOutboundConnections, 10)
	cfg.Node.MaxInboundConnections = util.NonZeroOrDefIn64(cfg.Node.MaxOutboundConnections, 115)

	// set connections hard limit
	if cfg.Node.MaxOutboundConnections > 10 {
		cfg.Node.MaxOutboundConnections = 10
	}
	if cfg.Node.MaxInboundConnections > 115 {
		cfg.Node.MaxInboundConnections = 115
	}

	// set to dev mode if -dev is set
	// and apply dev config values
	if devMode || debug {
		cfg.Node.Mode = config.ModeDev
		devDefaultConfig(cfg)
	}

	// check that the host address to bind
	// the engine to is valid,
	if !util.IsValidHostPortAddress(listeningAddr) {
		log.Fatal("invalid bind address provided")
	}

	// load the node account.
	// Required for signing blocks and transactions
	nodeKey, err := loadOrCreateAccount(account, password, seed)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Prevent mining when the node's key is ephemeral
	if mine && nodeKey.Meta["ephemeral"] != nil {
		log.Fatal(params.ErrMiningWithEphemeralKey.Error())
	}

	log.Info("Elld has started", "Version", cfg.VersionInfo.BuildVersion, "DevMode", devMode)

	// Create event the global event handler
	event := &emitter.Emitter{}

	// Create the local node.
	n, err := node.NewNode(cfg, listeningAddr, nodeKey, log)
	if err != nil {
		log.Fatal("failed to create local node", "Err", err.Error())
	}

	// In debug mode, we set log level
	// to DEBUG.
	if n.DevMode() {
		log.SetToDebug()
	}

	// Configure transactions pool and assign to node
	pool := txpool.New(params.PoolCapacity)
	pool.SetEventEmitter(event)
	n.SetTxsPool(pool)

	// Add hardcoded bootstrap addresses
	if err := n.AddAddresses(boostrapAddresses, true); err != nil {
		log.Fatal("%s", err)
	}

	// Add bootstrap addresses supplied
	// in the config file
	if err := n.AddAddresses(cfg.Node.BootstrapAddresses, false); err != nil {
		log.Fatal("%s", err)
	}

	// open the database on the engine
	if err = n.OpenDB(); err != nil {
		log.Fatal("failed to open local database", "Err", err.Error())
	}

	log.Info("Ready for connections", "Addr", n.GetAddress().ConnectionString())

	// Instantiate the blockchain manager,
	// set db, event emitter and pass it to the engine
	bchain := blockchain.New(n.GetTxPool(), cfg, log)
	bchain.SetDB(n.DB())
	bchain.SetEventEmitter(event)
	n.SetBlockchain(bchain)

	// power up the blockchain manager
	if err := bchain.Up(); err != nil {
		log.Fatal("failed to load blockchain manager", "Err", err.Error())
	}

	// Set the event handler in the node
	n.SetEventEmitter(event)

	// Start the node
	n.Start()

	// Initialized and start the miner if
	// enabled via the cli flag.
	miner := miner.NewMiner(nodeKey, bchain, event, cfg, log)
	miner.SetNumThreads(numMiners)
	if mine {
		go miner.Begin()
	}
	// Initialize and start the RPCServer
	// if enabled via the appropriate cli flag.
	var rpcServer = rpc.NewServer(n.DB(), rpcAddress, cfg, log)

	// Add the RPC APIs from various
	// components.
	rpcServer.AddAPI(
		n.APIs(),
		miner.APIs(),
		accountMgr.APIs(),
		bchain.APIs(),
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
		cs = console.New(nodeKey, consoleHistoryFilePath, cfg, log)
		cs.SetVersions(config.ProtocolVersion, BuildVersion, GoVersion, BuildCommit)
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

		profilePath := profile.ProfilePath(cfg.DataDir())
		cpuProfile, _ := cmd.Flags().GetBool("cpuprofile")
		if cpuProfile || os.Getenv("ELLD_CPU_PROFILING_ON") == "true" {
			defer profile.Start(profile.CPUProfile, profilePath).Stop()
		}

		memProfile, _ := cmd.Flags().GetBool("memprofile")
		if memProfile || os.Getenv("ELLD_MEM_PROFILING_ON") == "true" {
			defer profile.Start(profile.MemProfile, profilePath).Stop()
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
	startCmd.Flags().StringSliceP("addnode", "j", nil, "Add the address of a node to connect to.")
	startCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address local node will listen on.")
	startCmd.Flags().Bool("rpc", false, "Enables the RPC server")
	startCmd.Flags().String("rpcaddress", "127.0.0.1:8999", "Address RPC server will listen on.")
	startCmd.Flags().String("account", "", "The node's network account. The default account will be used if not set.")
	startCmd.Flags().String("pwd", "", "The password of the node's network account.")
	startCmd.Flags().Int64P("seed", "s", 0, "Provide a strong seed for network account creation (not recommended)")
	startCmd.Flags().Bool("mine", false, "Start proof-of-work mining")
	startCmd.Flags().Int("miners", 0, "The number of miner threads to use. (Default: Number of CPU)")
}

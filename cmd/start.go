package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/params"

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
// - If account is not provided, the default account is fetched and unlocked using
// 	 the password provided. If password is not set, it is requested via a prompt.
// - If account is not provided and no default account exists, an interactive account
// 	 creation session begins.
func loadOrCreateAccount(account, password string, seed int64) (*crypto.Key, error) {

	var address *crypto.Key
	var err error
	var storedAccount *accountmgr.StoredAccount

	if account != "" {
		if govalidator.IsNumeric(account) {
			aInt, err := strconv.Atoi(account)
			if err != nil {
				return nil, err
			}
			storedAccount, err = accountMgr.GetByIndex(aInt)
			if err != nil {
				return nil, err
			}
		} else {
			storedAccount, err = accountMgr.GetByAddress(account)
			if err != nil {
				return nil, err
			}
		}
	}

	if account == "" {
		storedAccount, err = accountMgr.GetDefault()
		if err != nil {
			return nil, fmt.Errorf("failed to get default account. %s", err)
		}
	}

	if storedAccount == nil {
		fmt.Println("No default account found. Create an account.")
		address, err = accountMgr.CreateCmd(seed, password)
		if err != nil {
			return nil, err
		}
	}

	if address != nil {
		return address, nil
	}

	// if address is unset, decrypt the account using the password provided.
	// if password is unset, request password from user
	// if password is set and is a path to a file, read the file and use its content as the password
	if password == "" {
		fmt.Println(fmt.Sprintf("Account {%s} needs to be unlocked. Please enter your password.", storedAccount.Address))
		password, err = accountMgr.AskForPasswordOnce()
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

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

	// When password is not set, get it from the
	// environment variable
	if len(password) == 0 {
		password = os.Getenv("ELLD_ACCOUNT_PASSWORD")
	}

	// Set configurations
	cfg.Node.MessageTimeout = util.NonZeroOrDefIn64(cfg.Node.MessageTimeout, 60)
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
	if devMode {
		cfg.Node.Mode = config.ModeDev
		devDefaultConfig(cfg)
	}

	// check that the host address to bind
	// the engine to is valid,
	if !util.IsValidHostPortAddress(listeningAddr) {
		log.Fatal("invalid bind address provided")
	}

	// load the coinbase account.
	// Required for signing blocks and transactions
	coinbase, err := loadOrCreateAccount(account, password, seed)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("Elld has started", "Version", cfg.VersionInfo.BuildVersion, "DevMode", devMode)

	// Create event the global event handler
	event := &emitter.Emitter{}

	// Create the local node.
	n, err := node.NewNode(cfg, listeningAddr, coinbase, log)
	if err != nil {
		log.Fatal("failed to create local node")
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
		log.Fatal("failed to open local database")
	}

	log.Info("Ready for connections", "Addr", n.GetAddress().ConnectionString())

	// Initialized gossip protocol handlers
	protocol := node.NewGossip(n, log)
	n.SetGossipProtocol(protocol)
	n.SetProtocolHandler(config.HandshakeVersion, protocol.OnHandshake)
	n.SetProtocolHandler(config.PingVersion, protocol.OnPing)
	n.SetProtocolHandler(config.GetAddrVersion, protocol.OnGetAddr)
	n.SetProtocolHandler(config.AddrVersion, protocol.OnAddr)
	n.SetProtocolHandler(config.IntroVersion, protocol.OnIntro)
	n.SetProtocolHandler(config.TxVersion, protocol.OnTx)
	n.SetProtocolHandler(config.BlockBodyVersion, protocol.OnBlockBody)
	n.SetProtocolHandler(config.RequestBlockVersion, protocol.OnRequestBlock)
	n.SetProtocolHandler(config.GetBlockHashesVersion, protocol.OnGetBlockHashes)
	n.SetProtocolHandler(config.GetBlockBodiesVersion, protocol.OnGetBlockBodies)

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
	miner := miner.New(coinbase, bchain, event, cfg, log)
	if mine {
		go miner.Mine()
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
		cs = console.New(coinbase, consoleHistoryFilePath, cfg, log)
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
}

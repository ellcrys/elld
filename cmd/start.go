package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ellcrys/ltcd/rpcclient"

	"github.com/ellcrys/elld/elldb"

	"github.com/ellcrys/elld/burner"

	"github.com/spf13/viper"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/ltcd"
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
func getKey(accountID, password string) (*crypto.Key, error) {

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

		// At this point, the user did not specify an account,
		// so we will just used the default account.
		storedAccount, err = accountMgr.GetDefault()
		if err != nil {
			if err == accountmgr.ErrAccountNotFound {
				return nil, fmt.Errorf(`No default account found. Node environment has ` +
					`not been initialized. Run 'elld init' to initialize the node or specify ` +
					`an existing account using '--account' flag.`)
			}
			return nil, err
		}
	}

	// If the password is unset, start an interactive session to
	// request password from user.
	if password == "" {
		fmt.Println(fmt.Sprintf("Account {%s} needs to be unlocked. Please enter your password.",
			storedAccount.Address))
		password, err = accountMgr.AskForPasswordOnce()
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	// If password is set and it's a path to a file,
	// read the file and use its content as the password
	if len(password) > 1 && (os.IsPathSeparator(password[0]) || (len(password) >= 2 &&
		password[:2] == "./")) {
		content, err := ioutil.ReadFile(password)
		if err != nil {
			if funk.Contains(err.Error(), "no such file") {
				return nil, fmt.Errorf("password file {%s} not found", password)
			}
			if funk.Contains(err.Error(), "is a directory") {
				return nil, fmt.Errorf("password file path {%s} is a directory. Expects a file",
					password)
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

	key := storedAccount.GetKey().(*crypto.Key)
	cfg.G().NodeKey = key

	return key, nil
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
	if viper.GetBool("burner.norpc") {
		args = append(args, "--norpc")
	}
	if viper.GetBool("burner.regtest") {
		args = append(args, "--regtest")
	}
	if miningAddr := viper.GetString("burner.miningaddr"); len(miningAddr) > 0 {
		args = append(args, []string{"--miningaddr", miningAddr}...)
	}
	if connect := viper.GetString("burner.connect"); len(connect) > 0 {
		args = append(args, []string{"--connect", connect}...)
	}
	if listen := viper.GetString("burner.listent"); len(listen) > 0 {
		args = append(args, []string{"--listen", listen}...)
	}
	return args
}

// startBurnerProcess sets up the burner server.
// It will return nil if the server stopped gracefully
// or error if it was interrupted.
func startBurnerProcess(cmd *cobra.Command, db elldb.DB, evt *emitter.Emitter) (chan error, *burner.API) {

	viper.BindPFlag("burner.testnet", cmd.Flags().Lookup("burner-testnet"))
	viper.BindPFlag("burner.listen", cmd.Flags().Lookup("burner-listen"))
	viper.BindPFlag("burner.rpcuser", cmd.Flags().Lookup("burner-rpcuser"))
	viper.BindPFlag("burner.rpcpass", cmd.Flags().Lookup("burner-rpcpass"))
	viper.BindPFlag("burner.notls", cmd.Flags().Lookup("burner-notls"))
	viper.BindPFlag("burner.rpclisten", cmd.Flags().Lookup("burner-rpclisten"))
	viper.BindPFlag("burner.utxokeeperskip", cmd.Flags().Lookup("burner-utxokeeperskip"))
	viper.BindPFlag("burner.utxokeeperworkers", cmd.Flags().Lookup("burner-utxokeeperworkers"))
	viper.BindPFlag("burner.utxokeeperoff", cmd.Flags().Lookup("burner-utxokeeperoff"))
	viper.BindPFlag("burner.utxokeeperreindex", cmd.Flags().Lookup("burner-utxokeeperreindex"))
	viper.BindPFlag("burner.utxokeeperfocus", cmd.Flags().Lookup("burner-utxokeeperfocus"))
	viper.BindPFlag("burner.norpc", cmd.Flags().Lookup("burner-norpc"))
	viper.BindPFlag("burner.regtest", cmd.Flags().Lookup("burner-regtest"))
	viper.BindPFlag("burner.miningaddr", cmd.Flags().Lookup("burner-miningaddr"))
	viper.BindPFlag("burner.connect", cmd.Flags().Lookup("burner-connect"))

	testnet := viper.GetBool("burner.testnet")
	noTLS := viper.GetBool("burner.notls")
	rpcUser := viper.GetString("burner.rpcuser")
	rpcPass := viper.GetString("burner.rpcpass")
	rpcListen := viper.GetString("burner.rpclisten")
	utxoKeeperSkip := viper.GetInt32("burner.utxokeeperskip")
	utxoKeeperNumThread := viper.GetInt("burner.utxokeeperworkers")
	noUTXOKeeper := viper.GetBool("burner.utxokeeperoff")
	reIndex := viper.GetBool("burner.utxokeeperreindex")
	focusAddr := viper.GetString("burner.utxokeeperfocus")

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

	var utxoKeeper *burner.UTXOIndexer
	var bc *rpcclient.Client
	// var blockWatcher *burner.BlockIndexer
	var tickman *burner.TicketManager
	var netVer = config.GetNetVersion()

	// We can start the UTXO keeper if --burner-utxokeeperoff is not set
	if noUTXOKeeper {
		log.Info("UTXO keeper is turned OFF")
		goto end
	}

	// When the burner rpc server is not enabled, do not proceed with starting
	// the UTXO keeper since it interacts with the burner chain via RPC.
	if !ltcd.IsRPCOn() {
		log.Warn("UTXO keeper is disabled because burner server RPC service is disabled")
		log.Warn("Block watcher is disabled because burner server RPC service is disabled")
		goto end
	}

	// At this point, the burner RPC service is enabled, so
	// we try to get a working client to it.
	bc, err = burner.GetClient(rpcListen, rpcUser, rpcPass, noTLS)
	if err != nil {
		log.Warn("Failed to connect to burner RPC service", "Err", err.Error())
		goto end
	}

	// Setup and start the UTXO keeper
	utxoKeeper = burner.NewUTXOIndexer(cfg, log, db, netVer, evt, interrupt)
	utxoKeeper.SetClient(bc)
	utxoKeeper.Begin(accountMgr, utxoKeeperNumThread, utxoKeeperSkip, reIndex, focusAddr)
	log.Info("UTXO keeper is turned ON")

	// Start the burner chain block watcher
	// blockWatcher = burner.NewBlockIndexer(cfg, log, db, evt, bc, interrupt)
	// blockWatcher.Begin()

	// Start the ticket manager
	tickman = burner.NewTicketManager(cfg, bc, interrupt)
	if err := tickman.Begin(); err != nil {
		log.Fatal("Failed to start ticket manager", "Err", err.Error())
	}

end:

	// Create a burner API
	burnerAPI := burner.NewAPI(db, cfg, evt, bc, accountMgr)

	return status, burnerAPI
}

// starts the node.
func start(cmd *cobra.Command, args []string, startConsole bool, interrupt chan struct{}) {

	var err error

	// Process flags
	viper.BindPFlag("node.id", cmd.Flags().Lookup("id"))
	viper.BindPFlag("node.password", cmd.Flags().Lookup("pwd"))
	viper.BindPFlag("node.bootstrapAddrs", cmd.Flags().Lookup("add-node"))
	viper.BindPFlag("node.address", cmd.Flags().Lookup("address"))
	viper.BindPFlag("rpc.enabled", cmd.Flags().Lookup("rpc"))
	viper.BindPFlag("rpc.address", cmd.Flags().Lookup("rpc-address"))
	viper.BindPFlag("rpc.disableAuth", cmd.Flags().Lookup("rpc-disable-auth"))
	viper.BindPFlag("rpc.sessionTTL", cmd.Flags().Lookup("rpc-session-ttl"))
	viper.BindPFlag("node.noNet", cmd.Flags().Lookup("no-net"))
	viper.BindPFlag("node.syncDisabled", cmd.Flags().Lookup("sync-disabled"))

	// Host chain config variables
	nodeID := viper.GetString("node.id")
	password := viper.GetString("node.password")
	listeningAddr := viper.GetString("node.address")
	startRPC := viper.GetBool("rpc.enabled")
	rpcAddress := viper.GetString("rpc.address")
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

	// Load the node key.
	// Required for signing blocks, transactions and for receiving mining rewards.
	nodeKey, err := getKey(nodeID, password)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create event the global event handler
	event := &emitter.Emitter{}

	// Create the local node.
	n, err := node.NewNode(cfg, listeningAddr, nodeKey, log)
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

	cfg.G().Bus = event
	cfg.G().DB = n.DB()

	// Initialize and set the blockchain manager's db,
	// event emitter and pass it to the engine
	bChain := blockchain.New(n.GetTxPool(), cfg, log)
	bChain.SetDB(n.DB())
	bChain.SetInterrupt(interrupt)
	bChain.SetEventEmitter(event)
	bChain.SetNodeKey(nodeKey)

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

	// Initialize the burner chain and related processes
	burnerStatus, burnerAPI := startBurnerProcess(cmd, n.DB(), event)

	// Start the block manager and the node
	n.Start()

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
		cs = console.New(nodeKey, consoleHistoryFilePath, cfg, log)
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
	Use:   "start [flags]",
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
	setStartFlags(startCmd)
}

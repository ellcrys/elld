package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/ellcrys/druid/rpc"

	"gopkg.in/asaskevich/govalidator.v4"

	"github.com/ellcrys/druid/accountmgr"
	funk "github.com/thoas/go-funk"

	"github.com/ellcrys/druid/configdir"
	"github.com/ellcrys/druid/console"
	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/node"
	"github.com/ellcrys/druid/util"
	"github.com/spf13/cobra"
)

var (
	hardcodedBootstrapNodes = []string{} // hardcoded bootstrap node address
)

func defaultConfig(cfg *configdir.Config) {
	cfg.Node.GetAddrInterval = util.NonZeroOrDefIn64(cfg.Node.GetAddrInterval, 10)
	cfg.Node.PingInterval = util.NonZeroOrDefIn64(cfg.Node.PingInterval, 60)
	cfg.Node.SelfAdvInterval = util.NonZeroOrDefIn64(cfg.Node.SelfAdvInterval, 10)
	cfg.Node.CleanUpInterval = util.NonZeroOrDefIn64(cfg.Node.CleanUpInterval, 10)
	cfg.Node.ConnEstInterval = util.NonZeroOrDefIn64(cfg.Node.ConnEstInterval, 10)
	cfg.TxPool.Capacity = util.NonZeroOrDefIn64(cfg.TxPool.Capacity, 100)
}

// loadAccount unlocks an account and returns the underlying address.
// - If account is provided, it is fetched and unlocked using the password provided.
//	 If password is not provided, the is requested through an interactive prompt.
// - If account is not provided, the default account is fetched and unlocked using
// 	 the password provided. If password is not set, it is requested via a prompt.
// - If account is not provided and no default account exists, an interactive account
// 	 creation session begins.
func loadAccount(account, password string) (*crypto.Key, error) {

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
		address, err = accountMgr.CreateCmd(password)
		if err != nil {
			return nil, err
		}
	}

	// if address is unset, decrypt the account using the password provided.
	// if password is unset, request password from user
	// if password is set and is a path to a file, read the file and use its content as the password
	if address == nil {

		if password == "" {
			fmt.Println(fmt.Sprintf("Account {%s} needs to be unlocked. Please enter your password.", storedAccount.Address))
			password, err = accountMgr.AskForPasswordOnce()
			if err != nil {
				log.Error(err.Error())
				return nil, err
			}
		}

		if len(password) > 0 && (os.IsPathSeparator(password[0]) || password[:2] == "./") {
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

		address = storedAccount.GetAddress()
	}

	return address, nil
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
// - start RPC server if enabled
// - start console if enabled
// - connect console to rpc server and prepare console vm if rpc server is enabled
func start(cmd *cobra.Command, args []string, startConsole bool) (*node.Node, *rpc.Server, *console.Console) {

	var err error

	bootstrapAddresses, _ := cmd.Flags().GetStringSlice("addnode")
	addressToListenOn, _ := cmd.Flags().GetString("address")
	startRPC, _ := cmd.Flags().GetBool("rpc")
	rpcAddress, _ := cmd.Flags().GetString("rpcaddress")
	account, _ := cmd.Flags().GetString("account")
	password, _ := cmd.Flags().GetString("pwd")

	if devMode {
		cfg.Node.Dev = devMode
		defaultConfig(cfg)
	}

	cfg.Node.MaxConnections = util.NonZeroOrDefIn64(cfg.Node.MaxConnections, 60)
	cfg.Node.BootstrapNodes = append(cfg.Node.BootstrapNodes, bootstrapAddresses...)
	cfg.Node.MaxAddrsExpected = 1000

	if !util.IsValidHostPortAddress(addressToListenOn) {
		log.Fatal("invalid bind address provided")
	}

	loadedAddress, err := loadAccount(account, password)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("Druid started", "Version", util.ClientVersion)

	n, err := node.NewNode(cfg, addressToListenOn, loadedAddress, log)
	if err != nil {
		log.Fatal("failed to create local node")
	}

	if n.DevMode() {
		log.SetToDebug()
	}

	if len(hardcodedBootstrapNodes) > 0 {
		if err := n.AddBootstrapNodes(hardcodedBootstrapNodes, true); err != nil {
			log.Fatal("%s", err)
		}
	}

	if len(cfg.Node.BootstrapNodes) > 0 {
		if err := n.AddBootstrapNodes(cfg.Node.BootstrapNodes, false); err != nil {
			log.Fatal("%s", err)
		}
	}

	if err = n.OpenDB(); err != nil {
		log.Fatal("failed to open local database")
	}

	log.Info("Waiting patiently to interact on", "Addr", n.GetMultiAddr(), "Dev", devMode)

	protocol := node.NewInception(n, log)
	n.SetProtocol(protocol)
	n.SetProtocolHandler(util.HandshakeVersion, protocol.OnHandshake)
	n.SetProtocolHandler(util.PingVersion, protocol.OnPing)
	n.SetProtocolHandler(util.GetAddrVersion, protocol.OnGetAddr)
	n.SetProtocolHandler(util.AddrVersion, protocol.OnAddr)

	n.Start()

	var rpcServer *rpc.Server
	if startRPC {
		rpcServer = rpc.NewServer(rpcAddress, n, log)
		go rpcServer.Serve()
	}

	var cs *console.Console
	if startConsole {

		cs = console.New(loadedAddress)

		if startRPC {

			err = cs.DialRPCServer(rpcAddress)
			if err != nil {
				log.Fatal("unable to start RPC server", "Err", err)
			}

			cs.PrepareVM()
			if err != nil {
				log.Fatal("unable to prepare console VM", "Err", err)
			}
		}

		fmt.Println("")
		go cs.Run()
	}

	return n, rpcServer, cs
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
  oldest account) in <CONFIGDIR>/` + configdir.AccountDirName + ` is used instead.
  
  If no account was found, an interactive session to create an account is started.   
  
  Account password will be interactively requested during account creation and unlock
  operations. Use '--pwd' flag to provide the account password non-interactively. '--pwd'
  can also accept a path to a file containing the password.`,
	Run: func(cmd *cobra.Command, args []string) {

		n, rpcServer, _ := start(cmd, args, false)

		onTerminate = func() {
			rpcServer.Stop()
			n.Stop()
		}

		n.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringSliceP("addnode", "j", nil, "IP of a node to connect to")
	startCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address local node will listen on")
	startCmd.Flags().Bool("rpc", false, "Launch RPC server")
	startCmd.Flags().String("rpcaddress", ":8999", "Address RPC server will listen on")
	startCmd.Flags().String("account", "", "Account to load. Default account is used if not provided")
	startCmd.Flags().String("pwd", "", "Used as password during initial account creation or loading an account")
}

package cmd

import (
	"github.com/ellcrys/druid/configdir"
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
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the node",
	Long:  `Start the node`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Druid started", "Version", util.ClientVersion)

		bootstrapAddresses, _ := cmd.Flags().GetStringSlice("addnode")
		addressToListenOn, _ := cmd.Flags().GetString("address")

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

		// create the local node
		n, err := node.NewNode(cfg, addressToListenOn, seed, log)
		if err != nil {
			log.Fatal("failed to create local node")
		}

		if n.DevMode() {
			log.SetToDebug()
		}

		// add hardcoded nodes
		if len(hardcodedBootstrapNodes) > 0 {
			if err := n.AddBootstrapNodes(hardcodedBootstrapNodes, true); err != nil {
				log.Fatal("%s", err)
			}
		}

		// add bootstrap nodes
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

		// set protocol and handlers
		n.SetProtocol(protocol)
		n.SetProtocolHandler(util.HandshakeVersion, protocol.OnHandshake)
		n.SetProtocolHandler(util.PingVersion, protocol.OnPing)
		n.SetProtocolHandler(util.GetAddrVersion, protocol.OnGetAddr)
		n.SetProtocolHandler(util.AddrVersion, protocol.OnAddr)

		// start the peer and cause main thread to wait
		n.Start()
		n.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringSliceP("addnode", "j", nil, "IP of a node to connect to")
	startCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address to listen on")
}

package cmd

import (
	"github.com/ellcrys/druid/configdir"
	"github.com/ellcrys/druid/peer"
	"github.com/ellcrys/druid/util"
	"github.com/spf13/cobra"
)

var (
	hardcodedBootstrapNodes = []string{} // hardcoded bootstrap node address
)

func defaultConfig(cfg *configdir.Config) {
	cfg.Peer.GetAddrInterval = util.NonZeroOrDefIn64(cfg.Peer.GetAddrInterval, 10)
	cfg.Peer.PingInterval = util.NonZeroOrDefIn64(cfg.Peer.PingInterval, 60)
	cfg.Peer.SelfAdvInterval = util.NonZeroOrDefIn64(cfg.Peer.SelfAdvInterval, 10)
	cfg.Peer.CleanUpInterval = util.NonZeroOrDefIn64(cfg.Peer.CleanUpInterval, 10)
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the peer",
	Long:  `Start the peer`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Druid started", "Version", util.ClientVersion)

		bootstrapAddresses, _ := cmd.Flags().GetStringSlice("addnode")
		addressToListenOn, _ := cmd.Flags().GetString("address")
		seed, _ := cmd.Flags().GetInt64("seed")
		dev, _ := cmd.Flags().GetBool("dev")

		if dev {
			cfg.Peer.Dev = dev
			defaultConfig(cfg)
		}

		cfg.Peer.MaxConnections = util.NonZeroOrDefIn64(cfg.Peer.MaxConnections, 60)
		cfg.Peer.BootstrapNodes = append(cfg.Peer.BootstrapNodes, bootstrapAddresses...)
		cfg.Peer.MaxAddrsExpected = 1000

		if !util.IsValidHostPortAddress(addressToListenOn) {
			log.Fatal("invalid bind address provided")
		}

		// create the peer
		p, err := peer.NewPeer(cfg, addressToListenOn, seed, log)
		if err != nil {
			log.Fatal("failed to create peer")
		}

		if p.DevMode() {
			log.SetToDebug()
		}

		// add hardcoded nodes
		if len(hardcodedBootstrapNodes) > 0 {
			if err := p.AddBootstrapPeers(hardcodedBootstrapNodes, true); err != nil {
				log.Fatal("%s", err)
			}
		}

		// add bootstrap nodes
		if len(cfg.Peer.BootstrapNodes) > 0 {
			if err := p.AddBootstrapPeers(cfg.Peer.BootstrapNodes, false); err != nil {
				log.Fatal("%s", err)
			}
		}

		log.Info("Waiting patiently to interact on", "Addr", p.GetMultiAddr(), "Dev", dev)

		protocol := peer.NewInception(p, log)

		// set protocol and handlers
		p.SetProtocol(protocol)
		p.SetProtocolHandler(util.HandshakeVersion, protocol.OnHandshake)
		p.SetProtocolHandler(util.PingVersion, protocol.OnPing)
		p.SetProtocolHandler(util.GetAddrVersion, protocol.OnGetAddr)
		p.SetProtocolHandler(util.AddrVersion, protocol.OnAddr)

		// start the peer and cause main thread to wait
		p.Start()
		p.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringSliceP("addnode", "j", nil, "IP of a node to connect to")
	startCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address to listen on")
	startCmd.Flags().Int64P("seed", "s", 0, "Random seed to use for identity creation")
	startCmd.Flags().Bool("dev", false, "Run client in development mode")
}

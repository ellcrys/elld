package cmd

import (
	"github.com/ellcrys/druid/configdir"
	"github.com/ellcrys/druid/peer"
	"github.com/ellcrys/druid/util"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	log                     *zap.SugaredLogger // logger
	hardcodedBootstrapNodes = []string{}       // hardcoded bootstrap node address
)

func init() {
	log = util.NewLogger("/peer")
}

// loadCfg loads the config file
func loadCfg(cfgDirPath string) (*configdir.Config, error) {

	cfgDir, err := configdir.NewConfigDir(cfgDirPath)
	if err != nil {
		return nil, err
	}

	if err := cfgDir.Init(); err != nil {
		return nil, err
	}

	cfg, err := cfgDir.Load()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the peer",
	Long:  `Start the peer`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infow("Druid has started", "Version", util.ClientVersion)

		bootstrapAddresses, _ := cmd.Flags().GetStringSlice("addnode")
		addressToListenOn, _ := cmd.Flags().GetString("address")
		seed, _ := cmd.Flags().GetInt64("seed")
		dev, _ := cmd.Flags().GetBool("dev")
		cfgDirPath, _ := cmd.Root().PersistentFlags().GetString("cfgdir")

		cfg, err := loadCfg(cfgDirPath)
		if err != nil {
			log.Fatal(err.Error())
		}

		cfg.Peer.BootstrapNodes = append(cfg.Peer.BootstrapNodes, bootstrapAddresses...)
		cfg.Peer.Dev = dev
		cfg.Peer.MaxAddrsExpected = 1000

		if cfg.Peer.MaxConnections == 0 {
			cfg.Peer.MaxConnections = 60
		}

		if !util.IsValidHostPortAddress(addressToListenOn) {
			log.Fatal("invalid bind address provided")
		}

		// create the peer
		log := util.NewLogger("peer")
		p, err := peer.NewPeer(cfg, addressToListenOn, seed, log)
		if err != nil {
			log.Fatalf("failed to create peer")
		}

		// add hardcoded nodes
		if len(hardcodedBootstrapNodes) > 0 {
			if err := p.AddBootstrapPeers(hardcodedBootstrapNodes, true); err != nil {
				log.Fatalf("%s", err)
			}
		}

		// add bootstrap nodes
		if len(cfg.Peer.BootstrapNodes) > 0 {
			if err := p.AddBootstrapPeers(cfg.Peer.BootstrapNodes, false); err != nil {
				log.Fatalf("%s", err)
			}
		}

		log.Infow("Waiting patiently to interact on", "Addr", p.GetMultiAddr(), "Dev", dev)

		protocol := peer.NewInception(p, log.Named("protocol"))

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
	rootCmd.PersistentFlags().String("cfgdir", "", "Configuration directory")
	startCmd.Flags().StringSliceP("addnode", "j", nil, "IP of a node to connect to")
	startCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address to listen on")
	startCmd.Flags().Int64P("seed", "s", 0, "Random seed to use for identity creation")
	startCmd.Flags().Bool("dev", false, "Run client in development mode")
}

package cmd

import (
	"github.com/ellcrys/gcoin/peer"
	"github.com/ellcrys/gcoin/util"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	log = util.NewLogger("/peer")
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the peer",
	Long:  `Start the peer`,
	Run: func(cmd *cobra.Command, args []string) {

		nodeToJoin, _ := cmd.Flags().GetStringSlice("addnode")
		addressToListenOn, _ := cmd.Flags().GetString("address")
		seed, _ := cmd.Flags().GetInt64("seed")

		if !util.IsValidHostPortAddress(addressToListenOn) {
			log.Fatal("invalid bind address provided")
		}

		// create the peer
		p, err := peer.NewPeer(addressToListenOn, seed)
		if err != nil {
			log.Fatalf("failed to create peer")
		}

		// add bootstrap nodes
		if len(nodeToJoin) > 0 {
			if err := p.AddBootstrapPeers(nodeToJoin); err != nil {
				log.Fatalf("%s", err)
			}
		}

		log.Infof("Node is listening at %s", p.GetMultiAddr())

		protocol := peer.NewProtocol(p)

		// set protocol handlers
		p.SetProtocolHandler(peer.HandshakeVersion, protocol.HandleHandshake)

		// start peer manager
		p.PM().Manage()

		// cause main thread to wait for peer
		p.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().StringSliceP("addnode", "j", nil, "IP of a node to connect to")
	startCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address to listen on")
	startCmd.Flags().Int64P("seed", "s", 0, "Random seed to use for identity creation")
}

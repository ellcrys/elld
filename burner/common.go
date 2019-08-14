package burner

import "github.com/ellcrys/elld/ltcsuite/ltcd/rpcclient"

// GetClient returns a client to the burner chain RPC server
func GetClient(host, rpcUser, rpcPass string, disableTLS bool) (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:       host,
		Endpoint:   "ws",
		User:       rpcUser,
		Pass:       rpcPass,
		DisableTLS: disableTLS,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

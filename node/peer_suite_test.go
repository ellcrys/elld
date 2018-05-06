package node

import (
	"os"
	path "path/filepath"
	"testing"

	"github.com/ellcrys/druid/configdir"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/ellcrys/druid/util/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log = logger.NewLogrusNoOp()
var cfg *configdir.Config

func setTestCfg() error {
	var err error
	dir, _ := homedir.Dir()
	cfgDir := path.Join(dir, ".ellcrys_test")
	os.MkdirAll(cfgDir, 0700)
	cfg, err = configdir.LoadCfg(cfgDir)
	cfg.Node.Dev = true
	cfg.Node.MaxAddrsExpected = 5
	cfg.Node.Test = true
	return err
}

func removeTestCfgDir() error {
	dir, _ := homedir.Dir()
	err := os.RemoveAll(path.Join(dir, ".ellcrys_test"))
	return err
}
func TestPeer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peer Suite")
}

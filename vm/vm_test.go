package vm

import (
	"fmt"
	"path/filepath"

	"github.com/ellcrys/druid/util/logger"
	"github.com/mitchellh/go-homedir"
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Vm", func() {

	hdir, _ := homedir.Dir()

	log := logger.NewLogrus()
	vm := New(log, filepath.Join(hdir, "mountdir"))
	fmt.Println(vm.Init())
})

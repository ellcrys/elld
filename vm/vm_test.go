package vm

import (
	"path/filepath"

	"github.com/ellcrys/druid/util/logger"
	"github.com/mitchellh/go-homedir"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vm", func() {
	hdir, _ := homedir.Dir()
	log := logger.NewLogrus()
	var vm *VM

	BeforeEach(func() {
		vm = New(log, filepath.Join(hdir, "mountdir"))
	})

	BeforeEach(func() {
		err := vm.Init()
		Expect(err).To(BeNil())
	})
})

package vm

import (
	"path/filepath"

	"github.com/ellcrys/elld/util/logger"
	"github.com/mitchellh/go-homedir"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VM", func() {

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

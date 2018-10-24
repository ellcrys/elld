package vm

import (
	"path/filepath"
	"testing"

	"github.com/ellcrys/elld/util/logger"
	"github.com/mitchellh/go-homedir"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestVM(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("VM", func() {
		hdir, _ := homedir.Dir()
		log := logger.NewLogrus()
		var vm *VM

		g.BeforeEach(func() {
			vm = New(log, filepath.Join(hdir, "mountdir"))
		})

		g.BeforeEach(func() {
			err := vm.Init()
			Expect(err).To(BeNil())
		})
	})
}

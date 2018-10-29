package miner

import (
	"testing"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestMiner(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
}

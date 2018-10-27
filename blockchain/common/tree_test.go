package common

import (
	"testing"

	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/merkletree"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestTree(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Tree", func() {

		g.Describe(".Add", func() {
			g.It("should add all items", func() {
				items := []merkletree.Content{TreeItem([]byte("a")), TreeItem([]byte("b"))}
				tree := NewTree()
				tree.Add(items[0])
				tree.Add(items[1])
				Expect(tree.items).To(HaveLen(2))
			})
		})

		g.Describe(".Build", func() {
			g.It("should return error if no item has been added", func() {
				tree := NewTree()
				err := tree.Build()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("error: cannot construct tree with no content"))
			})
		})

		g.Describe(".Root", func() {
			g.It("should return expected root", func() {
				items := []merkletree.Content{TreeItem([]byte("a")), TreeItem([]byte("b"))}
				tree := NewTree()
				tree.Add(items[0])
				tree.Add(items[1])
				Expect(tree.Build()).To(BeNil())
				Expect(tree.Root()).To(Equal(util.Hash{70, 13, 170, 184, 121, 200, 20, 163, 1, 149, 156, 20, 212, 181, 133, 63, 201, 200, 21, 159, 153, 118, 93, 88, 210, 135, 88, 77, 161, 255, 134, 58}))
			})
		})
	})
}

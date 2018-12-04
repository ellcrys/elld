package common

import (
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/merkletree"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Block", func() {

	Describe(".Add", func() {
		It("should add all items", func() {
			items := []merkletree.Content{TreeItem([]byte("a")), TreeItem([]byte("b"))}
			tree := NewTree()
			tree.Add(items[0])
			tree.Add(items[1])
			Expect(tree.items).To(HaveLen(2))
		})
	})

	Describe(".Build", func() {
		It("should return error if no item has been added", func() {
			tree := NewTree()
			err := tree.Build()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("error: cannot construct tree with no content"))
		})
	})

	Describe(".Root", func() {
		It("should return expected root", func() {
			items := []merkletree.Content{TreeItem([]byte("a")), TreeItem([]byte("b"))}
			tree := NewTree()
			tree.Add(items[0])
			tree.Add(items[1])
			Expect(tree.Build()).To(BeNil())
			Expect(tree.Root()).To(Equal(util.Hash{70, 13, 170, 184, 121, 200, 20, 163, 1, 149, 156, 20, 212, 181, 133, 63, 201, 200, 21, 159, 153, 118, 93, 88, 210, 135, 88, 77, 161, 255, 134, 58}))
		})
	})

})

package browscap

import (
	"github.com/1stborn/searchpattern"
)

type browserTree struct {
	rtree *searchpattern.RadixTree
}

func newTree() *browserTree {
	return &browserTree{
		rtree: searchpattern.CaseInsensitive(),
	}
}

func (bt *browserTree) Add(opts []string, br *Browser) {
	br.mapArray(opts)
	bt.rtree.Add(fPropertyName.GetString(opts), br)
}

func (bt *browserTree) Find(search string) *Browser {
	if b, ok := bt.rtree.FindFirst(search).(*Browser); ok {
		return b
	}

	return nil
}
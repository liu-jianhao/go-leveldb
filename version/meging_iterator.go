package version

import (
	"github.com/liu-jianhao/go-leveldb/memtable"
	"github.com/liu-jianhao/go-leveldb/sstable"
)

type MergingIterator struct {
	list    []*sstable.Iterator
	current *sstable.Iterator
}

func NewMergingIterator(list []*sstable.Iterator) *MergingIterator {
	return &MergingIterator{
		list: list,
	}
}

func (it *MergingIterator) Valid() bool {
	return it.current != nil && it.current.Valid()
}

func (it *MergingIterator) InternalKey() *memtable.InternalKey {
	return it.current.InternalKey()
}

func (it *MergingIterator) Next() {
	if it.current != nil {
		it.current.Next()
	}
	it.findSmallest()
}

func (it *MergingIterator) SeekToFirst() {
	for i := 0; i < len(it.list); i++ {
		it.list[i].SeekToFirst()
	}
	it.findSmallest()
}

func (it *MergingIterator) findSmallest() {
	var smallest *sstable.Iterator = nil
	for i := 0; i < len(it.list); i++ {
		if it.list[i].Valid() {
			if smallest == nil {
				smallest = it.list[i]
			} else if memtable.InternalKeyComparator(smallest.InternalKey(), it.list[i].InternalKey()) > 0 {
				smallest = it.list[i]
			}
		}
	}
	it.current = smallest
}

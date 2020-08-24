package memtable

import "github.com/liu-jianhao/go-leveldb/skiplist"

type Iterator struct {
	iter *skiplist.Iterator
}

func (it *Iterator) InternalKey() *InternalKey {
	return it.iter.Key().(*InternalKey)
}

func (it *Iterator) Valid() bool {
	return it.iter.Valid()
}

func (it *Iterator) Seek(target interface{}) {
	it.iter.Seek(target)
}

func (it *Iterator) SeekToFirst() {
	it.iter.SeekToFirst()
}

func (it *Iterator) SeekToLast() {
	it.iter.SeekToLast()
}

func (it *Iterator) Next() {
	it.iter.Next()
}

func (it *Iterator) Prev() {
	it.iter.Prev()
}

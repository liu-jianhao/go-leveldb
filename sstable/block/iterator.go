package block

import "github.com/liu-jianhao/go-leveldb/memtable"

type Iterator struct {
	block *Block
	index int
}

func (it *Iterator) Valid() bool {
	return it.index >= 0 && it.index < len(it.block.items)
}

func (it *Iterator) InternalKey() *memtable.InternalKey {
	return it.block.items[it.index]
}

func (it *Iterator) Next() {
	it.index++
}

func (it *Iterator) Prev() {
	it.index--
}

func (it *Iterator) Seek(target interface{}) {
	// 二分法查询
	left := 0
	right := len(it.block.items) - 1
	for left < right {
		mid := (left + right) / 2
		if memtable.UserKeyComparator(it.block.items[mid].UserKey, target) < 0 {
			left = mid + 1
		} else {
			right = mid
		}
	}
	if left == len(it.block.items)-1 {
		if memtable.UserKeyComparator(it.block.items[left].UserKey, target) < 0 {
			left++
		}
	}
	it.index = left
}

func (it *Iterator) SeekToFirst() {
	it.index = 0
}

func (it *Iterator) SeekToLast() {
	if len(it.block.items) > 0 {
		it.index = len(it.block.items) - 1
	}
}

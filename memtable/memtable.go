package memtable

import (
	"fmt"

	"github.com/liu-jianhao/go-leveldb/skiplist"
)

type MemTable struct {
	table *skiplist.SkipList
}

func NewMemTable() *MemTable {
	return &MemTable{
		table: skiplist.NewSkipList(InternalKeyComparator),
	}
}

func (m *MemTable) NewIterator() *Iterator {
	return &Iterator{
		iter: skiplist.NewIterator(m.table, nil),
	}
}

func (m *MemTable) Add(seq uint64, valueType ValueType, key, value []byte) {
	internalKey := &InternalKey{
		Seq:       seq,
		Type:      valueType,
		UserKey:   key,
		UserValue: value,
	}

	m.table.Insert(internalKey)
}

func (m *MemTable) Get(key []byte) ([]byte, error) {
	lookUpKey := LooUpKey(key)

	it := skiplist.NewIterator(m.table, nil)
	it.Seek(lookUpKey)
	if it.Valid() {
		internalKey := it.Key().(*InternalKey)
		if UserKeyComparator(key, internalKey.UserKey) == 0 {
			if internalKey.Type == TypeValue {
				return internalKey.UserValue, nil
			} else {
				return nil, fmt.Errorf("type deletion")
			}
		}
	}

	return nil, fmt.Errorf("not found")
}

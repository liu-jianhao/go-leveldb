package block

import (
	"bytes"
	"encoding/binary"

	"github.com/liu-jianhao/go-leveldb/memtable"
)

type Block struct {
	items []*memtable.InternalKey
}

func NewBlock(b []byte) *Block {
	data := bytes.NewBuffer(b)
	counter := binary.LittleEndian.Uint32(b[len(b)-4:])
	items := make([]*memtable.InternalKey, 0)

	for i := uint32(0); i < counter; i++ {
		var item memtable.InternalKey
		err := item.DecodeFrom(data)
		if err != nil {
			return nil
		}
		items = append(items, &item)
	}
	return &Block{
		items: items,
	}
}

func (block *Block) NewIterator() *Iterator {
	return &Iterator{block: block}
}

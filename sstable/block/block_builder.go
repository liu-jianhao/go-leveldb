package block

import (
	"bytes"
	"encoding/binary"

	"github.com/liu-jianhao/go-leveldb/memtable"
)

type BlockBuilder struct {
	buf     bytes.Buffer
	counter uint32
}

func (bb *BlockBuilder) Reset() {
	bb.counter = 0
	bb.buf.Reset()
}

func (bb *BlockBuilder) Add(item *memtable.InternalKey) error {
	bb.counter++
	return item.EncodeTo(&bb.buf)
}

func (bb *BlockBuilder) Finish() []byte {
	_ = binary.Write(&bb.buf, binary.LittleEndian, bb.counter)
	return bb.buf.Bytes()
}

func (bb *BlockBuilder) CurrentSizeEstimate() int {
	return bb.buf.Len()
}

func (bb *BlockBuilder) Empty() bool {
	return bb.buf.Len() == 0
}

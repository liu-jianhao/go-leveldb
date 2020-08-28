package sstable

import (
	"encoding/binary"

	"github.com/liu-jianhao/go-leveldb/memtable"
)

const (
	kTableMagicNumber uint64 = 0xdb4775248b80fb57
)

type BlockHandle struct {
	Offset uint32
	Size   uint32
}

func (blockHandle *BlockHandle) EncodeToBytes() []byte {
	p := make([]byte, 8)
	binary.LittleEndian.PutUint32(p, blockHandle.Offset)
	binary.LittleEndian.PutUint32(p[4:], blockHandle.Size)
	return p
}

func (blockHandle *BlockHandle) DecodeFromBytes(p []byte) {
	if len(p) == 8 {
		blockHandle.Offset = binary.LittleEndian.Uint32(p)
		blockHandle.Size = binary.LittleEndian.Uint32(p[4:])
	}
}

type IndexBlockHandle struct {
	*memtable.InternalKey
}

func (index *IndexBlockHandle) SetBlockHandle(blockHandle BlockHandle) {
	index.UserValue = blockHandle.EncodeToBytes()
}

func (index *IndexBlockHandle) GetBlockHandle() (blockHandle BlockHandle) {
	blockHandle.DecodeFromBytes(index.UserValue)
	return
}

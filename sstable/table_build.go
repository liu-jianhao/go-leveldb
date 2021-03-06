package sstable

import (
	"os"

	"github.com/liu-jianhao/go-leveldb/memtable"
	"github.com/liu-jianhao/go-leveldb/sstable/block"
)

const (
	MAX_BLOCK_SIZE = 4 * 1024
)

type TableBuilder struct {
	file               *os.File
	offset             uint32
	numEntries         int32
	dataBlockBuilder   block.BlockBuilder
	indexBlockBuilder  block.BlockBuilder
	pendingIndexEntry  bool
	pendingIndexHandle IndexBlockHandle
	status             error
}

func NewTableBuilder(fileName string) *TableBuilder {
	var builder TableBuilder
	var err error
	builder.file, err = os.Create(fileName)
	if err != nil {
		return nil
	}
	builder.pendingIndexEntry = false
	return &builder
}

func (builder *TableBuilder) FileSize() uint32 {
	return builder.offset
}

func (builder *TableBuilder) Add(internalKey *memtable.InternalKey) {
	if builder.status != nil {
		return
	}
	if builder.pendingIndexEntry {
		_ = builder.indexBlockBuilder.Add(builder.pendingIndexHandle.InternalKey)
		builder.pendingIndexEntry = false
	}
	// todo : filter block

	builder.pendingIndexHandle.InternalKey = internalKey

	builder.numEntries++
	_ = builder.dataBlockBuilder.Add(internalKey)
	if builder.dataBlockBuilder.CurrentSizeEstimate() > MAX_BLOCK_SIZE {
		builder.flush()
	}
}
func (builder *TableBuilder) flush() {
	if builder.dataBlockBuilder.Empty() {
		return
	}
	orgKey := builder.pendingIndexHandle.InternalKey
	builder.pendingIndexHandle.InternalKey = memtable.NewInternalKey(orgKey.Seq, orgKey.Type, orgKey.UserKey, nil)
	builder.pendingIndexHandle.SetBlockHandle(builder.writeBlock(&builder.dataBlockBuilder))
	builder.pendingIndexEntry = true
}

func (builder *TableBuilder) Finish() error {
	// write data block
	builder.flush()
	// todo : filter block

	// write index block
	if builder.pendingIndexEntry {
		_ = builder.indexBlockBuilder.Add(builder.pendingIndexHandle.InternalKey)
		builder.pendingIndexEntry = false
	}
	var footer Footer
	footer.IndexHandle = builder.writeBlock(&builder.indexBlockBuilder)

	// write footer block
	_ = footer.EncodeTo(builder.file)
	_ = builder.file.Close()
	return nil
}

func (builder *TableBuilder) writeBlock(blockBuilder *block.BlockBuilder) BlockHandle {
	content := blockBuilder.Finish()
	// todo : compress, crc
	var blockHandle BlockHandle
	blockHandle.Offset = builder.offset
	blockHandle.Size = uint32(len(content))
	builder.offset += uint32(len(content))
	_, builder.status = builder.file.Write(content)
	_ = builder.file.Sync()
	blockBuilder.Reset()
	return blockHandle
}

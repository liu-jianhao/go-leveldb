package sstable

import (
	"fmt"
	"io"
	"os"

	"github.com/liu-jianhao/go-leveldb/memtable"
	"github.com/liu-jianhao/go-leveldb/sstable/block"
)

type SsTable struct {
	index *block.Block
	footer Footer
	file *os.File
}

func Open(fileName string) (*SsTable, error) {
	var table SsTable
	var err error
	table.file, err = os.Open(fileName)
	if err != nil {
		return nil, err
	}

	stat, _ := table.file.Stat()
	footerSize := int64(table.footer.Size())
	if stat.Size() < footerSize {
		return nil, fmt.Errorf("file too small")
	}

	_, err = table.file.Seek(-footerSize, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	err = table.footer.DecodeFrom(table.file)
	if err != nil {
		return nil, err
	}

	table.index = table.readBlock(table.footer.IndexHandle)
	return &table, nil
}

func (table *SsTable) Get(key []byte) ([]byte, error) {
	it := NewIterator()
	it.Seek(key)
	if it.Valid() {
		internalKey := it.InternalKey()
		if memtable.UserKeyComparator(key,internalKey) == 0{
			return memtable.UserValue, nil
		} else {
			return nil, fmt.Errorf("not deletion")
		}
	}
	return nil, fmt.Errorf("not found")
}

func (table  *SsTable) readBlock(blockHandle BlockHandle) *block.Block {
	p := make([]byte, blockHandle.Size)
	n, err := table.file.ReadAt(p, int64(blockHandle.Offset))
	if err != nil || uint32(n) != blockHandle.Size {
		return nil
	}

	return block.New(p)
}
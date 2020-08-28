package block

import (
	"testing"

	"github.com/liu-jianhao/go-leveldb/memtable"
)

func Test_SsTable(t *testing.T) {
	var builder BlockBuilder

	item := memtable.NewInternalKey(1, memtable.TypeValue, []byte("123"), []byte("1234"))
	_ = builder.Add(item)
	item = memtable.NewInternalKey(2, memtable.TypeValue, []byte("124"), []byte("1245"))
	_ = builder.Add(item)
	item = memtable.NewInternalKey(3, memtable.TypeValue, []byte("125"), []byte("0245"))
	_ = builder.Add(item)
	b := builder.Finish()

	block := NewBlock(b)
	it := block.NewIterator()

	it.Seek([]byte("1244"))
	if it.Valid() {
		if string(it.InternalKey().UserKey) != "125" {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

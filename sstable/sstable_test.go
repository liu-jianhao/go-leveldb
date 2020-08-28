package sstable

import (
	"fmt"
	"testing"

	"github.com/liu-jianhao/go-leveldb/memtable"
)

func Test_SsTable(t *testing.T) {
	builder := NewTableBuilder("D:\\000123.ldb")
	item := memtable.NewInternalKey(1, memtable.TypeValue, []byte("123"), []byte("1234"))
	builder.Add(item)
	item = memtable.NewInternalKey(2, memtable.TypeValue, []byte("124"), []byte("1245"))
	builder.Add(item)
	item = memtable.NewInternalKey(3, memtable.TypeValue, []byte("125"), []byte("0245"))
	builder.Add(item)
	_ = builder.Finish()

	table, err := Open("D:\\000123.ldb")
	fmt.Println(err)
	if err == nil {
		fmt.Println(table.index)
		fmt.Println(table.footer)
	}
	it := table.NewIterator()
	it.Seek([]byte("1244"))
	if it.Valid() {
		if string(it.InternalKey().UserKey) != "125" {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

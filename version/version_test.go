package version

import (
	"fmt"
	"testing"

	"github.com/liu-jianhao/go-leveldb/memtable"
)

const (
	dbName1 = "TestGet"
	dbName2 = "TestLoad"
)

func Test_Version_Get(t *testing.T) {
	v := NewVersion(dbName1)
	var f FileMetaData
	f.number = 123
	f.smallest = memtable.NewInternalKey(1, memtable.TypeValue, []byte("123"), nil)
	f.largest = memtable.NewInternalKey(1, memtable.TypeValue, []byte("125"), nil)
	v.files[0] = append(v.files[0], &f)

	value, err := v.Get([]byte("125"))
	fmt.Println(err, value)
}

func Test_Version_Load(t *testing.T) {
	v := NewVersion(dbName2)
	memTable := memtable.NewMemTable()
	memTable.Add(1234567, memtable.TypeValue, []byte("aadsa34a"), []byte("bb23b3423"))
	v.WriteLevel0Table(memTable)
	n, _ := v.Save()
	fmt.Println(v)

	v2, _ := Load(dbName2, n)
	fmt.Println(v2)
	value, err := v2.Get([]byte("aadsa34a"))
	fmt.Println(err, value)
}

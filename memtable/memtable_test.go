package memtable

import (
	"fmt"
	"testing"
)

func Test_MemTable(t *testing.T) {
	memTable := NewMemTable()
	memTable.Add(1234567, TypeValue, []byte("aadsa34a"), []byte("bb23b3423"))
	value, _ := memTable.Get([]byte("aadsa34a"))
	fmt.Println(string(value))
}

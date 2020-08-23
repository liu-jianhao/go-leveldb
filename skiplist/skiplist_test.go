package skiplist

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	skipList := NewSkipList(IntComparator)
	for i := 0; i < 10; i++ {
		rand.Seed(time.Now().UnixNano())
		skipList.Insert(rand.Int() % 10)
	}
	it := NewIterator(skipList, nil)
	for it.SeekToFirst(); it.Valid(); it.Next() {
		fmt.Printf("%v ", it.Key())
	}
	fmt.Println()
	for it.SeekToLast(); it.Valid(); it.Prev() {
		fmt.Printf("%v ", it.Key())
	}
	fmt.Println()
}

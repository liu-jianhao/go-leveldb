package version

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/liu-jianhao/go-leveldb/sstable"
	"github.com/liu-jianhao/go-leveldb/utils"
)

const (
	NumCache = 1000
)

type TableCache struct {
	lock   sync.Mutex
	dbName string
	cache  *lru.Cache
}

func NewTableCache(name string) *TableCache {
	cache, _ := lru.New(NumCache)
	return &TableCache{
		dbName: name,
		cache:  cache,
	}
}

func (tc *TableCache) NewIterator(fileNum uint64) *sstable.Iterator {
	table, _ := tc.findTable(fileNum)
	if table != nil {
		return table.NewIterator()
	}
	return nil
}

func (tc *TableCache) Get(fileNum uint64, key []byte) ([]byte, error) {
	table, err := tc.findTable(fileNum)
	if table != nil {
		return table.Get(key)
	}
	return nil, err
}

func (tc *TableCache) Evict(fileNum uint64) {
	tc.cache.Remove(fileNum)
}

func (tc *TableCache) findTable(fileNum uint64) (*sstable.SsTable, error) {
	tc.lock.Lock()
	defer tc.lock.Unlock()

	table, ok := tc.cache.Get(fileNum)
	if ok {
		return table.(*sstable.SsTable), nil
	} else {
		st, err := sstable.Open(utils.TableFileName(tc.dbName, fileNum))
		tc.cache.Add(fileNum, st)
		return st, err
	}
}

package skiplist

type Iterator struct {
	skipList *SkipList
	node     *Node
}

func NewIterator(list *SkipList, node *Node) *Iterator {
	return &Iterator{
		skipList: list,
		node:     node,
	}
}

func (it *Iterator) Valid() bool {
	return it.node != nil
}

func (it *Iterator) Key() interface{} {
	return it.node.key
}

func (it *Iterator) Next() {
	it.skipList.lock.Lock()
	defer it.skipList.lock.Unlock()

	it.node = it.node.Next(0)
}

func (it *Iterator) Prev() {
	it.skipList.lock.Lock()
	defer it.skipList.lock.Unlock()

	it.node = it.skipList.FindLessThan(it.node.key)
	if it.node == it.skipList.head {
		it.node = nil
	}
}

func (it *Iterator) Seek(target interface{}) {
	it.skipList.lock.Lock()
	defer it.skipList.lock.Unlock()

	it.node, _ = it.skipList.FindGreaterOrEqual(target)
}

func (it *Iterator) SeekToFirst() {
	it.skipList.lock.Lock()
	defer it.skipList.lock.Unlock()

	it.node = it.skipList.head.Next(0)
}

func (it *Iterator) SeekToLast() {
	it.skipList.lock.Lock()
	defer it.skipList.lock.Unlock()

	it.node = it.skipList.FindLast()
	if it.node == it.skipList.head {
		it.node = nil
	}
}

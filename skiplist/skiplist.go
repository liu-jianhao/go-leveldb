package skiplist

import (
	"math/rand"
	"sync"
)

const (
	kMaxHeight = 12
)

type SkipList struct {
	comparator Comparator
	head       *Node
	maxHeight  int
	lock       sync.Mutex
}

func NewSkipList(comparator Comparator) *SkipList {
	return &SkipList{
		comparator: comparator,
		head:       NewNode(nil, kMaxHeight),
		maxHeight:  1,
	}
}

// 先找到不小于该key的node（FindGreaterOrEqual（）），随机产生新node的高度，对各个高度的链表做insert即可。
func (sl *SkipList) Insert(key interface{}) {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	_, prev := sl.FindGreaterOrEqual(key)
	height := sl.RandomHeight() // 随机获取一个 level 值
	if height > sl.maxHeight {
		for i := sl.maxHeight; i < height; i++ {
			prev[i] = sl.head
		}
		sl.maxHeight = height
	}

	// 生成节点并插入
	x := NewNode(key, height)
	for i := 0; i < height; i++ {
		x.SetNext(i, prev[i].Next(i))
		prev[i].SetNext(i, x)
	}
}

func (sl *SkipList) Equal(a, b interface{}) bool {
	return sl.comparator(a, b) == 0
}

func (sl *SkipList) Contains(key interface{}) bool {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	x, _ := sl.FindGreaterOrEqual(key)
	if x != nil && sl.Equal(key, x.key) {
		return true
	}
	return false
}

func (sl *SkipList) RandomHeight() int {
	height, kBranching := 1, 4
	for height < kMaxHeight && (rand.Intn(kBranching) == 0) {
		height++
	}
	return height
}

func (sl *SkipList) KeyIsAfterNode(key interface{}, node *Node) bool {
	return node != nil && (sl.comparator(node.key, key) < 0)
}

func (sl *SkipList) FindGreaterOrEqual(key interface{}) (*Node, [kMaxHeight]*Node) {
	var prev [kMaxHeight]*Node
	x := sl.head              // 从头结点开始查找
	level := sl.maxHeight - 1 // 从最高层开始查找
	for {
		next := x.Next(level)
		if sl.KeyIsAfterNode(key, next) {
			x = next // 待查找 key 比 next 大，则在该层继续查找
		} else {
			prev[level] = x
			if level == 0 {
				return next, prev
			} else {
				level--
			}
		}
	}
}

func (sl *SkipList) FindLessThan(key interface{}) *Node {
	x := sl.head
	level := sl.maxHeight - 1
	for {
		next := x.Next(level)
		if next == nil || sl.comparator(next.key, key) >= 0 {
			if level == 0 {
				return x
			} else {
				level--
			}
		} else {
			x = next
		}
	}
}

func (sl *SkipList) FindLast() *Node {
	x := sl.head
	level := sl.maxHeight - 1
	for {
		next := x.Next(level)
		if next == nil {
			if level == 0 {
				return x
			} else {
				level--
			}
		} else {
			x = next
		}
	}
}

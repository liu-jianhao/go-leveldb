# 背景
leveldb是一个google实现的非常高效的kv数据库，能够支持billion级别的数据量。
https://github.com/google/leveldb

## 准备工作
首先要了解leveldb的基本架构，下面我就列几个参考资料：
https://yuerblog.cc/wp-content/uploads/leveldb%E5%AE%9E%E7%8E%B0%E8%A7%A3%E6%9E%90.pdf

## 开始搬砖
### skiplist
+ 类似BigTable的模式，数据在内存中以**memtable**形式存储。leveldb的**memtable**实现没有使用复杂的B树系列，采用的是更轻量级的**skip list**。
+ 全局看来，**skip list**所有的node就是一个排序的链表，考虑到操作效率，为这一个链表再添加若干不同跨度的辅助链表，查找时通过辅助链表可以跳跃比较来加大查找的步进。
+ 每个链表上都是排序的node，而每个node也可能同时处在多个链表上。将一个node所属链表的数量看作它的高度，那么，不同高度的node在查找时会获得不同跳跃跨度的查找优化。
+ 换个角度，如果node的高度具有随机性，数据集合从高度层次上看就有了散列性，也就等同于树的平衡。
+ 相对于其他树型数据结构采用不同策略来保证平衡状态，**skip list**仅保证新加入node的高度随机即可（当然也可以采用规划计算的方式确定高度，以获得平摊复杂度。leveldb采用的是更简单的随机方式）

**skiplist**的操作：
1. 写入（SkipList::Insert()/Delete()）
    + `insert`: 先找到不小于该key的node（`FindGreaterOrEqual()`），随机产生新node的高度，对各个高度的链表做insert即可。
    + `delete`: 先找到node，并对其所在各个高度的链表做相应的更新。leveldb中delete操作相当于insert，skiplist代码中并未实现。
2. 读取**skiplist**提供了Iterator的接口方式，供查找和遍历时使用。
    + `Seek()`找到不小key的节点（`FindGreaterOrEqual()`）。从根节点开始，高度从高向低与node的key比较，直到找到或者到达链表尾。
    + `SeekToFirst()`定位到头节点最低高度的node即可。
    + `SeekToLast()`从头节点的最高开始，依次前进，知道达到链表尾。
    + `Next()/Prev()`在最低高度的链表上做next或者prev即可。
    
### memtable
**memtable**的操作：
1. 写入（`MemTable::Add()`）
    + 将传入的key和value dump成**memtable**中存储的数据格式。
    + `SkipList::Insert()`。
2. 读取(`MemTable::Get()`）**memtable**对key的查找和遍历封装成MemTableIterator。底层直接使用SkipList的类Iterator接口。
    + 从传入的LookupKey中取得**memtable**中存储的key格式。
    + 做`MemTableIterator::Seek()`。
    + seek失败，返回data not exist。seek成功，则判断数据的ValueType
        + kTypeValue，返回对应的value数据。
        + kTypeDeletion，返回data not exist。

### block
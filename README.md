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
1. 写入（BlockBuilder::Add()/BlockBuilder::Finish()）
    + block写入时,不会对key做排序的逻辑，因为sstable的产生是由memtable dump或者compact时merge 排序产生，key的顺序上层已经保证。
    + 检查上一轮前缀压缩是否已经完成（达到restart_interval）完成，则记录restarts点，重新开始新一轮。该key不做任何处理（shared_bytes = 0）未完成，计算该key与保存的上一个key的相同前缀，确定unshared_bytes/shared_bytes
    + 将key/value 以block内entry的数据格式，追加到该block上(内存中)。
    + BlockBuilder::Finish()在一个block完成（达到设定的block_size）时，将restarts点的集合和数量追加到block上。
2. 读取（ReadBlock() table/format.cc）有了一个block的BlockHandle，即可定位到该block在sstable中的offset及size,从而读取出具体的block（ReadBlock()）。
    + 根据BlockHandle，将block从sstable中读取出来（包含trailer）。
    + 可选校验trailer中的crc(get时由ReadOption::verify_checksums控制，compact时由Option::paranoid_checks控制)。
    + 根据trailer中的type，决定是否要解压数据。d.将数据封装成Block（block.cc），解析出restarts集合以及数量。
    
    
### sstable
sstable是leveldb中持久化数据的文件格式。整体来看，sstable由数据(data)和元信息(meta/index)组成。数据和元信息统一以block单位存储（除了文件最末尾的footer元信息），读取时也采用统一的读取逻辑。


1. 写入（TableBuilder::Add() TableBuilder::Finish()）同sstable中block的写入一样，不需要关心排序。
    + 如果是一个新block的开始，计算出上一个block的end-key（Comparator::FindShortestSeparator()），连同BlockHandle添加到index_block中。
    考虑到index_block会load进内存，为了节约index_block中保存的index信息（每个block对应的end-key/offset/size），
    leveldb中并没有直接使用block最后一个key做为它的end-key，而是使用Comparator::FindShortestSeparator（）得到。
    默认实现是将大于上一个block最后一个key，但小于下一个block第一个key的最小key作为上一个block的end-key。用户可以实现自己的Comparator来控制这个策略。
    + 将key/value 加入当前data_block（BlockBuilder::Add()）。
    + 如果当前data_block达到设定的Option::block_size，将data_block写入磁盘(BlockBuilder::WriteBlock()）。
    + BlockBuilder::Finish()。e.对block的数据做可选的压缩（snppy），append到sstable文件。
    + 添加该block的trailer（type/crc），append到sstable文件。
    + 记录该block的BlockHandle。
    + TableBuilder::Finish()是在sstable完成时（dump memtable完成或者达到kTargetFileSize）做的处理。
        + 将meta_index_block写入磁盘（当前未实现meta_index_block逻辑，meta_index_block没有任何数据）。
        + 计算最后一个block的end-key（Comparator::FindShortSuccessor()），连同其BlockHandle添加到index_block中。
        + 将index_block写入磁盘。
        + 构造footer，作为最后部分写入sstable
2. 读取(Table::Open() table/table.cc TwoLevelIteratortable/two_level_iterator.cc)一个sstable需要IO时首先open(Table::Open()).
    + 根据传入的sstable size（Version::files_保存的FileMetaData），首先读取文件末尾的footer。
    + 解析footer数据(Footer::DecodeFrom()table/format.cc)，校验magic,获得index_block和metaindex_block的BlockHandle.
    + 根据index_block的BlockHandle，读取index_block(ReadBlock()table/format.cc)。
    + 分配cacheID(ShardedLRUCache::NewId(), util/cache.cc)。
    + 封装成Table（调用者会将其加入table cache，TableCache::NewIterator（））。对sstable进行key的查找遍历封装成TwoLevelIterator(参见Iterator)处理。
3. cache的处理（TableCache db/table_cache.cc）加快block的定位，对sstable的元信息做了cache(TableCache),使用ShardLRUCache。
    + cache的key为sstable的FileNumber，value是封装了元信息的Table句柄。每当新加入TableCache时，会获得一个全局唯一cacheId。
    当compact完成，删除sstable文件的同时，会从TableCache中将其对应的entry清除。
    而属于该sstable的BlockCache可能有多个，需要遍历BlockCache才能得到（或者构造sstable中所有block的BlockCache的key做查询），所以基于效率考虑，
    BlockCache中属于该sstable的block缓存entry并不做处理，由BlockCache的LRU逻辑自行清除。
4. 统一处理cache与IO （TableCache::NewIterator（）db/table_cache.cc）处理table cache和实际sstable IO的逻辑由TableCache::NewIterator（）控制。
    + 构造table cache中的key（FileNumber）,对TableCache做Lookup，若存在，则直接获得对应的Table。若不存在，则根据FileNumber构造出sstable的具体路径，Table::Open()，得到具体的Table,并插入TableCache。
    + 返回sstable的Iterator（Table::NewIterator()，TwoLevelIterator）。上层对sstable进行key的查找遍历都是用TableCache::NewIterator（）获得sstable的Iterator,然后做后续操作，无需关心cache相关逻辑。
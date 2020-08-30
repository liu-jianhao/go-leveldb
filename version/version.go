package version

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"sort"

	"github.com/liu-jianhao/go-leveldb/memtable"
	"github.com/liu-jianhao/go-leveldb/sstable"
	"github.com/liu-jianhao/go-leveldb/utils"
)

const (
	NumLevel = 7
)

type Version struct {
	tableCache     *TableCache
	nextFileNumber uint64
	seq            uint64
	files          [NumLevel][]*FileMetaData
	compactPointer [NumLevel]*memtable.InternalKey
}

func NewVersion(name string) *Version {
	return &Version{
		tableCache:     NewTableCache(name),
		nextFileNumber: 1,
	}
}

func Load(dbName string, number uint64) (*Version, error) {
	fileName := utils.DescriptorFileName(dbName, number)
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	v := NewVersion(dbName)
	return v, v.DecodeFrom(file)
}

func (v *Version) Save() (uint64, error) {
	tmp := v.nextFileNumber
	fileName := utils.DescriptorFileName(v.tableCache.dbName, v.nextFileNumber)
	v.nextFileNumber++
	file, err := os.Create(fileName)
	if err != nil {
		return tmp, err
	}
	defer file.Close()
	return tmp, v.EncodeTo(file)
}

func (v *Version) Log() {
	for level := 0; level < NumLevel; level++ {
		for i := 0; i < len(v.files[level]); i++ {
			log.Printf("version[%d]: %d", level, v.files[level][i].number)
		}
	}
}

func (v *Version) Copy() *Version {
	var c Version

	c.tableCache = v.tableCache
	c.nextFileNumber = v.nextFileNumber
	c.seq = v.seq
	for level := 0; level < NumLevel; level++ {
		c.files[level] = make([]*FileMetaData, len(v.files[level]))
		copy(c.files[level], v.files[level])
	}
	return &c
}

func (v *Version) NextSeq() uint64 {
	v.seq++
	return v.seq
}

func (v *Version) NumLevelFiles(l int) int {
	return len(v.files[l])
}

func (v *Version) Get(key []byte) ([]byte, error) {
	var tmp []*FileMetaData
	var tmp2 [1]*FileMetaData
	var files []*FileMetaData
	// We can search level-by-level since entries never hop across
	// levels.  Therefore we are guaranteed that if we find data
	// in an smaller level, later levels are irrelevant.
	for level := 0; level < NumLevel; level++ {
		numFiles := len(v.files[level])
		if numFiles == 0 {
			continue
		}
		if level == 0 {
			// Level-0 files may overlap each other.  Find all files that
			// overlap user_key and process them in order from newest to oldest.
			for i := 0; i < numFiles; i++ {
				f := v.files[level][i]
				if memtable.UserKeyComparator(key, f.smallest.UserKey) >= 0 && memtable.UserKeyComparator(key, f.largest.UserKey) <= 0 {
					tmp = append(tmp, f)
				}
			}
			if len(tmp) == 0 {
				continue
			}
			sort.Slice(tmp, func(i, j int) bool { return tmp[i].number > tmp[j].number })
			numFiles = len(tmp)
			files = tmp
		} else {
			index := v.findFile(v.files[level], key)
			if index >= numFiles {
				files = nil
				numFiles = 0
			} else {
				tmp2[0] = v.files[level][index]
				if memtable.UserKeyComparator(key, tmp2[0].smallest.UserKey) < 0 {
					files = nil
					numFiles = 0
				} else {
					files = tmp2[:]
					numFiles = 1
				}
			}
		}
		for i := 0; i < numFiles; i++ {
			f := files[i]
			value, err := v.tableCache.Get(f.number, key)
			if err != utils.ErrNotFound {
				return value, err
			}
		}
	}
	return nil, utils.ErrNotFound
}

func (v *Version) findFile(files []*FileMetaData, key []byte) int {
	left := 0
	right := len(files)
	for left < right {
		mid := (left + right) / 2
		f := files[mid]
		if memtable.UserKeyComparator(f.largest.UserKey, key) < 0 {
			// Key at "mid.largest" is < "target".  Therefore all
			// files at or before "mid" are uninteresting.
			left = mid + 1
		} else {
			// Key at "mid.largest" is >= "target".  Therefore all files
			// after "mid" are uninteresting.
			right = mid
		}
	}
	return right
}

func (v *Version) EncodeTo(w io.Writer) error {
	_ = binary.Write(w, binary.LittleEndian, v.nextFileNumber)
	_ = binary.Write(w, binary.LittleEndian, v.seq)
	for level := 0; level < NumLevel; level++ {
		numFiles := len(v.files[level])
		_ = binary.Write(w, binary.LittleEndian, int32(numFiles))

		for i := 0; i < numFiles; i++ {
			_ = v.files[level][i].EncodeTo(w)
		}
	}
	return nil
}

func (v *Version) DecodeFrom(r io.Reader) error {
	_ = binary.Read(r, binary.LittleEndian, &v.nextFileNumber)
	_ = binary.Read(r, binary.LittleEndian, &v.seq)
	var numFiles int32
	for level := 0; level < NumLevel; level++ {
		_ = binary.Read(r, binary.LittleEndian, &numFiles)
		v.files[level] = make([]*FileMetaData, numFiles)
		for i := 0; i < int(numFiles); i++ {
			var meta FileMetaData
			_ = meta.DecodeFrom(r)
			v.files[level][i] = &meta
		}
	}
	return nil
}

func (v *Version) deleteFile(level int, meta *FileMetaData) {
	numFiles := len(v.files[level])
	for i := 0; i < numFiles; i++ {
		if v.files[level][i].number == meta.number {
			v.files[level] = append(v.files[level][:i], v.files[level][i+1:]...)
			log.Printf("deleteFile, level:%d, num:%d", level, meta.number)
			break
		}
	}
}

func (v *Version) addFile(level int, meta *FileMetaData) {
	log.Printf("addFile, level:%d, num:%d, %s-%s", level, meta.number, string(meta.smallest.UserKey), string(meta.largest.UserKey))
	if level == 0 {
		// 0层没有排序
		v.files[level] = append(v.files[level], meta)
	} else {
		numFiles := len(v.files[level])
		index := v.findFile(v.files[level], meta.smallest.UserKey)
		if index >= numFiles {
			v.files[level] = append(v.files[level], meta)
		} else {
			var tmp []*FileMetaData
			tmp = append(tmp, v.files[level][:index]...)
			tmp = append(tmp, meta)
			v.files[level] = append(tmp, v.files[level][index:]...)
		}
	}
}

func (v *Version) WriteLevel0Table(imm *memtable.MemTable) {
	var meta FileMetaData
	meta.allowSeeks = 1 << 30
	meta.number = v.nextFileNumber
	v.nextFileNumber++
	builder := sstable.NewTableBuilder(utils.TableFileName(v.tableCache.dbName, meta.number))
	iter := imm.NewIterator()
	iter.SeekToFirst()
	if iter.Valid() {
		meta.smallest = iter.InternalKey()
		for ; iter.Valid(); iter.Next() {
			meta.largest = iter.InternalKey()
			builder.Add(iter.InternalKey())
		}
		_ = builder.Finish()
		meta.fileSize = uint64(builder.FileSize())
		meta.smallest.UserValue = nil
		meta.largest.UserValue = nil
	}

	// 挑选合适的level
	level := 0
	if !v.overlapInLevel(0, meta.smallest.UserKey, meta.largest.UserKey) {
		for ; level < 2; level++ {
			if v.overlapInLevel(level+1, meta.smallest.UserKey, meta.largest.UserKey) {
				break
			}
		}
	}

	v.addFile(level, &meta)
}

func (v *Version) overlapInLevel(level int, smallestKey, largestKey []byte) bool {
	numFiles := len(v.files[level])
	if numFiles == 0 {
		return false
	}
	if level == 0 {
		for i := 0; i < numFiles; i++ {
			f := v.files[level][i]
			if memtable.UserKeyComparator(smallestKey, f.largest.UserKey) > 0 || memtable.UserKeyComparator(f.smallest.UserKey, largestKey) > 0 {
				continue
			} else {
				return true
			}
		}
	} else {
		index := v.findFile(v.files[level], smallestKey)
		if index >= numFiles {
			return false
		}
		if memtable.UserKeyComparator(largestKey, v.files[level][index].smallest.UserKey) > 0 {
			return true
		}
	}
	return false
}

func (v *Version) DoCompactionWork() bool {
	c := v.pickCompaction()
	if c == nil {
		return false
	}
	log.Printf("DoCompactionWork begin\n")
	defer log.Printf("DoCompactionWork end\n")
	c.Log()
	if c.isTrivialMove() {
		// Move file to next level
		v.deleteFile(c.level, c.inputs[0][0])
		v.addFile(c.level+1, c.inputs[0][0])
		return true
	}
	var list []*FileMetaData
	var current_key *memtable.InternalKey
	iter := v.makeInputIterator(c)
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		var meta FileMetaData
		meta.allowSeeks = 1 << 30
		meta.number = v.nextFileNumber
		v.nextFileNumber++
		builder := sstable.NewTableBuilder(utils.TableFileName(v.tableCache.dbName, meta.number))

		meta.smallest = iter.InternalKey()
		for ; iter.Valid(); iter.Next() {
			if current_key != nil {
				// 去除重复的记录
				ret := memtable.UserKeyComparator(iter.InternalKey().UserKey, current_key.UserKey)
				if ret == 0 {
					continue
				} else if ret < 0 {
					log.Fatalf("%s < %s", string(iter.InternalKey().UserKey), string(current_key.UserKey))
				}
				current_key = iter.InternalKey()
			}
			meta.largest = iter.InternalKey()
			builder.Add(iter.InternalKey())
			if builder.FileSize() > 2<<20 {
				break
			}
		}
		_ = builder.Finish()
		meta.fileSize = uint64(builder.FileSize())
		meta.smallest.UserValue = nil
		meta.largest.UserValue = nil

		list = append(list, &meta)
	}

	for i := 0; i < len(c.inputs[0]); i++ {
		v.deleteFile(c.level, c.inputs[0][i])
	}
	for i := 0; i < len(c.inputs[1]); i++ {
		v.deleteFile(c.level+1, c.inputs[1][i])
	}
	for i := 0; i < len(list); i++ {
		v.addFile(c.level+1, list[i])
	}
	return true
}

func (v *Version) makeInputIterator(c *Compaction) *MergingIterator {
	var list []*sstable.Iterator
	for i := 0; i < len(c.inputs[0]); i++ {
		list = append(list, v.tableCache.NewIterator(c.inputs[0][i].number))
	}
	for i := 0; i < len(c.inputs[1]); i++ {
		list = append(list, v.tableCache.NewIterator(c.inputs[1][i].number))
	}
	return NewMergingIterator(list)
}

func (v *Version) pickCompaction() *Compaction {
	var c Compaction
	c.level = v.pickCompactionLevel()
	if c.level < 0 {
		return nil
	}
	var smallest, largest *memtable.InternalKey
	// Files in level 0 may overlap each other, so pick up all overlapping ones
	if c.level == 0 {
		c.inputs[0] = append(c.inputs[0], v.files[c.level]...)
		smallest = c.inputs[0][0].smallest
		largest = c.inputs[0][0].largest
		for i := 1; i < len(c.inputs[0]); i++ {
			f := c.inputs[0][i]
			if memtable.InternalKeyComparator(f.largest, largest) > 0 {
				largest = f.largest
			}
			if memtable.InternalKeyComparator(f.smallest, smallest) < 0 {
				smallest = f.smallest
			}
		}
	} else {
		// Pick the first file that comes after compact_pointer_[level]
		for i := 0; i < len(v.files[c.level]); i++ {
			f := v.files[c.level][i]
			if v.compactPointer[c.level] == nil || memtable.InternalKeyComparator(f.largest, v.compactPointer[c.level]) > 0 {
				c.inputs[0] = append(c.inputs[0], f)
				break
			}
		}
		if len(c.inputs[0]) == 0 {
			c.inputs[0] = append(c.inputs[0], v.files[c.level][0])
		}
		smallest = c.inputs[0][0].smallest
		largest = c.inputs[0][0].largest
	}

	for i := 0; i < len(v.files[c.level+1]); i++ {
		f := v.files[c.level+1][i]

		if memtable.InternalKeyComparator(f.largest, smallest) < 0 || memtable.InternalKeyComparator(f.smallest, largest) > 0 {
			// "f" is completely before specified range; skip it,  // "f" is completely after specified range; skip it
		} else {
			c.inputs[1] = append(c.inputs[1], f)
		}
	}
	return &c
}

func (v *Version) pickCompactionLevel() int {
	compactionLevel := -1
	bestScore := 1.0
	score := 0.0
	for level := 0; level < NumLevel-1; level++ {
		if level == 0 {
			score = float64(len(v.files[0])) / float64(4)
		} else {
			score = float64(totalFileSize(v.files[level])) / maxBytesForLevel(level)
		}

		if score > bestScore {
			bestScore = score
			compactionLevel = level
		}

	}
	return compactionLevel
}

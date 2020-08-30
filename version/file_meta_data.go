package version

import (
	"encoding/binary"
	"io"

	"github.com/liu-jianhao/go-leveldb/memtable"
)

type FileMetaData struct {
	allowSeeks uint64
	number     uint64
	fileSize   uint64
	smallest   *memtable.InternalKey
	largest    *memtable.InternalKey
}

func (fmd *FileMetaData) EncodeTo(w io.Writer) error {
	_ = binary.Write(w, binary.LittleEndian, fmd.allowSeeks)
	_ = binary.Write(w, binary.LittleEndian, fmd.fileSize)
	_ = binary.Write(w, binary.LittleEndian, fmd.number)
	_ = fmd.smallest.EncodeTo(w)
	_ = fmd.largest.EncodeTo(w)
	return nil
}
func (fmd *FileMetaData) DecodeFrom(r io.Reader) error {
	_ = binary.Read(r, binary.LittleEndian, &fmd.allowSeeks)
	_ = binary.Read(r, binary.LittleEndian, &fmd.fileSize)
	_ = binary.Read(r, binary.LittleEndian, &fmd.number)
	fmd.smallest = new(memtable.InternalKey)
	_ = fmd.smallest.DecodeFrom(r)
	fmd.largest = new(memtable.InternalKey)
	_ = fmd.largest.DecodeFrom(r)
	return nil
}

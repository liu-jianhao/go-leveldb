package memtable

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

type ValueType int8

const (
	TypeDeletion ValueType = 0
	TypeValue    ValueType = 1
)

type InternalKey struct {
	Seq       uint64
	Type      ValueType
	UserKey   []byte
	UserValue []byte
}

func NewInternalKey(seq uint64, valueType ValueType, key, value []byte) *InternalKey {
	return &InternalKey{
		Seq:       seq,
		Type:      valueType,
		UserKey:   key,
		UserValue: value,
	}
}

func (ik *InternalKey) EncodeTo(w io.Writer) error {
	_ = binary.Write(w, binary.LittleEndian, ik.Seq)
	_ = binary.Write(w, binary.LittleEndian, ik.Type)
	_ = binary.Write(w, binary.LittleEndian, int32(len(ik.UserKey)))
	_ = binary.Write(w, binary.LittleEndian, ik.UserKey)
	_ = binary.Write(w, binary.LittleEndian, int32(len(ik.UserValue)))
	return binary.Write(w, binary.LittleEndian, ik.UserValue)
}

func (ik *InternalKey) DecodeFrom(r io.Reader) error {
	var tmp int32
	_ = binary.Read(r, binary.LittleEndian, &ik.Seq)
	_ = binary.Read(r, binary.LittleEndian, &ik.Type)
	_ = binary.Read(r, binary.LittleEndian, &tmp)
	ik.UserKey = make([]byte, tmp)
	_ = binary.Read(r, binary.LittleEndian, ik.UserKey)
	_ = binary.Read(r, binary.LittleEndian, &tmp)
	ik.UserValue = make([]byte, tmp)
	return binary.Read(r, binary.LittleEndian, ik.UserValue)
}

func LooUpKey(key []byte) *InternalKey {
	return NewInternalKey(math.MaxUint64, TypeValue, key, nil)
}

func InternalKeyComparator(a, b interface{}) int {
	aKey, bKey := a.(*InternalKey), b.(*InternalKey)
	r := UserKeyComparator(aKey.UserKey, bKey.UserKey)
	if r == 0 {
		aNum := aKey.Seq
		bNum := bKey.Seq
		if aNum > bNum {
			r = -1
		} else if aNum < bNum {
			r = 1
		}
	}
	return r
}

func UserKeyComparator(a, b interface{}) int {
	aKey, bKey := a.([]byte), b.([]byte)
	return bytes.Compare(aKey, bKey)
}

package sstable

import (
	"encoding/binary"
	"io"
)

type Footer struct {
	MetaIndexHandle BlockHandle
	IndexHandle     BlockHandle
}

func (footer *Footer) Size() int {
	// add magic size
	return binary.Size(footer) + 8
}

func (footer *Footer) EncodeTo(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, footer)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, kTableMagicNumber)
	return err
}

func (footer *Footer) DecodeFrom(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, footer)
	if err != nil {
		return err
	}
	var magic uint64
	err = binary.Read(r, binary.LittleEndian, &magic)
	if err != nil {
		return err
	}
	if magic != kTableMagicNumber {
		return fmt.Errorf("error magic number")
	}
	return nil
}


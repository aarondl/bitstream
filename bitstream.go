package bitstream

import (
	"bufio"
	"errors"
	"io"
)

// Bitstream reads many different types of values outside byte alignments.
type Bitstream struct {
	reader *bufio.Reader

	offset uint
	bits   byte
}

// NewBitstream wraps the reader in a bufio.Reader.
func NewBitstream(reader io.Reader) *Bitstream {
	return &Bitstream{
		offset: 8,
		reader: bufio.NewReader(reader),
	}
}

// Bits returns the next bits up to a max of 64.
func (b *Bitstream) Bits(nBits int) (val uint64, err error) {
	if nBits > 64 {
		panic("Next can only pull back 64 bits at a time.")
	}

	var bitOffset uint
	for nBits > 0 {
		if b.offset == 8 {
			b.offset = 0
			b.bits, err = b.reader.ReadByte()
			if err != nil {
				return val, err
			}
		}

		toRead := uint(nBits)
		if toRead > (8 - b.offset) {
			toRead = 8 - b.offset
		}

		var mask byte = ((1 << toRead) - 1) << b.offset

		val |= (uint64(mask&b.bits) >> b.offset) << bitOffset
		bitOffset += toRead
		b.offset += toRead
		nBits -= int(toRead)
	}

	return val, nil
}

// Bytes from the reader.
func (b *Bitstream) Bytes(dst []byte) error {
	for i := 0; i < len(dst); i++ {
		bits, err := b.Bits(8)
		if err != nil {
			return err
		}

		dst[i] = byte(bits & 0xFF)
	}

	return nil
}

// BitsInBytes returns the number of requested bits inside a byte array.
func (b *Bitstream) BitsInBytes(dst []byte, nBits int) (err error) {
	var byteOffset int
	var bitOffset uint

	if len(dst) < (nBits+7)/8 {
		return errors.New("bitsinbytes: buffer too small")
	}

	for nBits > 0 {
		if b.offset == 8 {
			b.offset = 0

			b.bits, err = b.reader.ReadByte()
			if err != nil {
				return err
			}
		}

		if bitOffset == 8 {
			bitOffset = 0
			byteOffset++
		}

		maskSize := uint(nBits)
		if maskSize > (8 - bitOffset) {
			maskSize = 8 - bitOffset
		}
		if maskSize > (8 - b.offset) {
			maskSize = 8 - b.offset
		}
		if maskSize > 8 {
			maskSize = 8
		}

		var mask byte = ((1 << maskSize) - 1) << b.offset

		dst[byteOffset] |= ((mask & b.bits) >> b.offset) << bitOffset
		bitOffset += maskSize
		b.offset += maskSize
		nBits -= int(maskSize)
	}

	return nil
}

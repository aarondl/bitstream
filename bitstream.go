// Package bitstream is used to read bits out of any given reader.
// It can operate in regular or "shift up" mode which determines how
// reads across byte-alignment construct values (with the current or the next
// byte able to become the most significant bits). It also has optimized
// fast-paths on byte alignments.
package bitstream

import (
	"bufio"
	"errors"
	"io"
)

var (
	bufferTooSmall = errors.New("bitsinbytes: buffer too small")
)

// Reader reads many different types of values outside byte alignments.
type Reader struct {
	reader *bufio.Reader

	offset uint
	bits   byte

	// Bits returns the next bits up to a max of 64.
	Bits func(nBits int) (val uint64, err error)
	// Bytes returns the number of requested bits inside a byte array.
	Bytes func(dst []byte, nBits int) (err error)
}

// New constructs a reader that shifts the next byte up to become
// the most significant bits. Given data: 1010 0000 | 0000 0101,
// a read of Bits(16) will yield: 0000 0101 1010 0000
func New(reader io.Reader) *Reader {
	r := &Reader{
		offset: 8,
		reader: bufio.NewReader(reader),
	}

	r.Bits = r.bitsLow
	r.Bytes = r.bytesLow

	return r
}

// NewShiftUp constructs a reader that shifts the current byte up to become
// the most significant bits. Given data: 1010 0000 | 0000 0101,
// a read of Bits(16) will yield: 1010 0000 0000 0101
func NewShiftUp(reader io.Reader) *Reader {
	r := &Reader{
		offset: 8,
		reader: bufio.NewReader(reader),
	}

	r.Bits = r.bitsHigh
	r.Bytes = r.bytesHigh

	return r
}

// Align discards the rest of the current byte's bits and byte-aligns the reader.
func (r *Reader) Align() {
	r.offset = 8
}

func (r *Reader) bitsLow(nBits int) (val uint64, err error) {
	if nBits > 64 {
		panic("Can only read 64 bits at a time.")
	}

	var bitOffset uint
	for nBits > 0 {
		if r.offset == 8 {
			r.offset = 0
			r.bits, err = r.reader.ReadByte()
			if err != nil {
				return val, err
			}
		}

		toRead := uint(nBits)
		if toRead > (8 - r.offset) {
			toRead = 8 - r.offset
		}

		var mask byte = ((1 << toRead) - 1) << r.offset

		val |= (uint64(mask&r.bits) >> r.offset) << bitOffset
		bitOffset += toRead
		r.offset += toRead
		nBits -= int(toRead)
	}

	return val, nil
}

// Byte from the reader.
func (r *Reader) Byte() (byte, error) {
	if r.offset == 8 {
		return r.reader.ReadByte()
	}

	bits, err := r.Bits(8)
	return byte(bits), err
}

// Read whole bytes from the reader.
func (r *Reader) Read(dst []byte) (int, error) {
	if r.offset == 8 {
		ret, err := r.reader.Read(dst)

		// bufio doesn't fill it's buffer until it's completely empty.
		// if a short read happens with no error: retry.
		if err == nil && len(dst) != ret {
			again, e := r.reader.Read(dst[ret:])
			return again + ret, e
		}

		return ret, err
	}

	n := 0
	for i := 0; i < len(dst); i++ {
		bits, err := r.Bits(8)
		if err != nil {
			return n, err
		}

		dst[i] = byte(bits & 0xFF)
		n++
	}

	return n, nil
}

func (r *Reader) bytesLow(dst []byte, nBits int) (err error) {
	var byteOffset int
	var bitOffset uint

	if len(dst) < (nBits+7)/8 {
		return bufferTooSmall
	}

	for nBits > 0 {
		if r.offset == 8 {
			r.offset = 0

			r.bits, err = r.reader.ReadByte()
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
		if maskSize > (8 - r.offset) {
			maskSize = 8 - r.offset
		}

		var mask byte = ((1 << maskSize) - 1) << r.offset

		dst[byteOffset] |= ((mask & r.bits) >> r.offset) << bitOffset
		bitOffset += maskSize
		r.offset += maskSize
		nBits -= int(maskSize)
	}

	return nil
}

func (r *Reader) bitsHigh(nBits int) (val uint64, err error) {
	if nBits > 64 {
		panic("Can only read 64 bits at a time.")
	}

	var bitOffset uint
	for nBits > 0 {
		if r.offset == 8 {
			r.offset = 0
			r.bits, err = r.reader.ReadByte()
			if err != nil {
				return val, err
			}
		}

		toRead := uint(nBits)
		if toRead > (8 - r.offset) {
			toRead = 8 - r.offset
		}

		val = (val << toRead) | uint64(r.bits>>r.offset)&((1<<toRead)-1)
		bitOffset += toRead
		r.offset += toRead
		nBits -= int(toRead)
	}

	return val, nil
}

func (r *Reader) bytesHigh(dst []byte, nBits int) (err error) {
	var byteOffset int
	var bitOffset uint

	if len(dst) < (nBits+7)/8 {
		return bufferTooSmall
	}

	for nBits > 0 {
		if r.offset == 8 {
			r.offset = 0

			r.bits, err = r.reader.ReadByte()
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
		if maskSize > (8 - r.offset) {
			maskSize = 8 - r.offset
		}

		dst[byteOffset] = (dst[byteOffset] << maskSize) | byte(r.bits>>r.offset)&((1<<maskSize)-1)
		bitOffset += maskSize
		r.offset += maskSize
		nBits -= int(maskSize)
	}

	return nil
}

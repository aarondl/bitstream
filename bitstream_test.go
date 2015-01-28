package bitstream

import (
	"bytes"
	"io"
	"testing"
)

var reader io.Reader = &Reader{}

func toBinInt(s string) uint64 {
	var val uint64
	var offset uint
	for i := len(s) - 1; i >= 0; i-- {
		switch s[i] {
		case '1':
			val |= 1 << offset
			offset++
		case '0':
			offset++
		case ' ':
		default:
			panic("Not a valid binary number.")
		}
	}

	return val
}

func toBin(s string) byte {
	return byte(toBinInt(s))
}

func TestReader_Bits(t *testing.T) {
	data := []byte{toBin("0000 1111"), toBin("1010 0101"), toBin("1111 0000")}

	b := New(bytes.NewBuffer(data))
	var val uint64
	var err error

	if val, err = b.Bits(5); err != nil {
		t.Error("Unexpected Error:", err)
	} else if val != toBinInt("01111") {
		t.Errorf("Wrong Value: % 02X", val)
	}
	if val, err = b.Bits(6); err != nil {
		t.Error("Unexpected Error:", err)
	} else if val != toBinInt("101000") {
		t.Errorf("Wrong Value: % 02X", val)
	}
	if val, err = b.Bits(1); err != nil {
		t.Error("Unexpected Error:", err)
	} else if val != toBinInt("0") {
		t.Errorf("Wrong Value: % 02X", val)
	}
	if val, err = b.Bits(3); err != nil {
		t.Error("Unexpected Error:", err)
	} else if val != toBinInt("010") {
		t.Errorf("Wrong Value: % 02X", val)
	}
	if val, err = b.Bits(2); err != nil {
		t.Error("Unexpected Error:", err)
	} else if val != toBinInt("01") {
		t.Errorf("Wrong Value: % 02X", val)
	}
	if val, err = b.Bits(7); err != nil {
		t.Error("Unexpected Error:", err)
	} else if val != toBinInt("1111 000") {
		t.Errorf("Wrong Value: % 02X", val)
	}
	if val, err = b.Bits(4); err == nil {
		t.Error("Expected an error to occur on overflow.")
	} else if val != 0 {
		t.Error("Expected 0 value.")
	}
}

func TestReader_Byte(t *testing.T) {
	data := []byte{0xF0, 0xFF, 0x0F}

	b := New(bytes.NewBuffer(data))

	if _, err := b.Bits(4); err != nil {
		t.Error(err)
	}

	if bits, err := b.Byte(); err != nil {
		t.Error(err)
	} else if bits != 0xFF {
		t.Errorf("Value was not correct: % 02X", bits)
	}
	if bits, err := b.Byte(); err != nil {
		t.Error(err)
	} else if bits != 0xFF {
		t.Errorf("Value was not correct: % 02X", bits)
	}

	if _, err := b.Bits(4); err != nil {
		t.Error(err)
	}
}

func TestReader_Read(t *testing.T) {
	data := []byte{0xF0, 0xFF, 0x0F}

	b := New(bytes.NewBuffer(data))

	if _, err := b.Bits(4); err != nil {
		t.Error(err)
	}

	val := make([]byte, 2)
	if n, err := b.Read(val); err != nil {
		t.Error(err)
	} else if n != 2 {
		t.Error("Number of bytes wrong:", n)
	}

	if val[0] != 0xFF || val[1] != 0xFF {
		t.Errorf("The values were not correctly read: % 02X - % 02X", val[0], val[1])
	}

	if _, err := b.Bits(4); err != nil {
		t.Error(err)
	}
}

func TestReader_Bytes(t *testing.T) {
	data := []byte{0x00, 0xF0, 0xFF, 0x0F, 0x00}

	b := New(bytes.NewBuffer(data))
	var err error

	val := make([]byte, 2)
	if err = b.Bytes(val, 12); err != nil {
		t.Error("Unexpected Error:", err)
	} else if bytes.Compare([]byte{0x00, 0x00}, val) != 0 {
		t.Errorf("Wrong Value: % 02X", val)
	}

	val = make([]byte, 3)
	if err = b.Bytes(val, 22); err != nil {
		t.Error("Unexpected Error:", err)
	} else if bytes.Compare([]byte{0xFF, 0xFF, 0x00}, val) != 0 {
		t.Errorf("Wrong Value: % 02X", val)
	}

	val = make([]byte, 1)
	if err = b.Bytes(val, 6); err != nil {
		t.Error("Unexpected Error:", err)
	} else if bytes.Compare([]byte{0x00}, val) != 0 {
		t.Errorf("Wrong Value: % 02X", val)
	}

	if err = b.Bytes(val, 6); err == nil {
		t.Error("Expected an overflow error.")
	}
}

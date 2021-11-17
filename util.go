package util

/*
#cgo CFLAGS: -Wno-deprecated -std=c99

#include <string.h>

typedef unsigned char * uint8ptr;
typedef short * int16ptr;
typedef short int16;
typedef unsigned short uint16;

void convert_to_int16(uint8ptr src, size_t srcLen, int16ptr dst, size_t dstLen) {
    if (srcLen <= (dstLen * 2)) {
        uint8ptr tmp = (uint8ptr)dst;
        memcpy(tmp, src, srcLen);
    }
}
*/
import "C"

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"reflect"
	"strconv"
	"unsafe"
)

// Atou16 convert a string to uint16
func Atou16(s string) uint16 {
	return uint16(Atoi(s))
}

// Atou32 convert a string to uint32
func Atou32(s string) uint32 {
	return uint32(Atoi(s))
}

// Atoi convert a string to int
func Atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		buf := []byte(s)
		for k := range buf {
			if buf[k] < '0' || buf[k] > '9' {
				i, _ = strconv.Atoi(string(buf[0:k]))
				break
			}
		}
	}
	return i
}

// Itoa convert int to a string
func Itoa(i int) string {
	return strconv.Itoa(i)
}

// This is a faster implentation of strconv.Itoa() without importing another library
// https://stackoverflow.com/a/39444005
func I32toa(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

// uint16base64 is a faster version of fmt.Sprintf("0x%04x", n)
//
// BenchmarkUint16Base16/fmt.Sprintf-8         	10000000	       154 ns/op	       8 B/op	       2 allocs/op
// BenchmarkUint16Base16/uint16base16-8        	50000000	        35.0 ns/op	       8 B/op	       1 allocs/op
func U16base16(n uint16) string {
	var digit16 = []byte("0123456789abcdef")
	b := []byte("0x0000")
	b[5] = digit16[n&0x000f]
	b[4] = digit16[n&0x00f0>>4]
	b[3] = digit16[n&0x0f00>>8]
	b[2] = digit16[n&0xf000>>12]
	return string(b)
}

// ValueToBytes convert a uint16/uint32/uint64(Little-Endian) to []byte.
func ValueToBytes(T interface{}) []byte {
	size := reflect.TypeOf(T).Size()
	if size != 2 && size != 4 && size != 8 {
		return nil
	}

	bytes := make([]byte, size)
	if size == 2 {
		binary.LittleEndian.PutUint16(bytes, T.(uint16))
	} else if size == 4 {
		binary.LittleEndian.PutUint32(bytes, T.(uint32))
	} else if size == 8 {
		binary.LittleEndian.PutUint64(bytes, T.(uint64))
	} else {
		return nil
	}
	return bytes
}
func Uint16ToBytes(val uint16) []byte {
	return ValueToBytes(val)
}
func Uint32ToBytes(val uint32) []byte {
	return ValueToBytes(val)
}

// BytesToValue convert []byte to a uint16/uint32/uint64(Little-Endian)
func BytesToValue(bytes []byte) interface{} {
	size := len(bytes)
	if size == 2 {
		return binary.LittleEndian.Uint16(bytes)
	} else if size == 4 {
		return binary.LittleEndian.Uint32(bytes)
	} else if size == 8 {
		return binary.LittleEndian.Uint64(bytes)
	} else {
		return 0
	}
}
func BytesToUint16(bytes []byte) uint16 {
	return BytesToValue(bytes).(uint16)
}
func BytesToUint32(bytes []byte) uint32 {
	return BytesToValue(bytes).(uint32)
}

// ValueOrderChange convert a uint16/uint32/uint64(LittleEndian/BigEndian) to
// another uint16/uint32/uint64(BigEndian/LittleEndian).
func ValueOrderChange(T interface{}, order binary.ByteOrder) interface{} {
	bytes := ValueToBytes(T)
	if bytes == nil {
		log.Println("invalid bytes in ValueOrderChange")
		return 0
	}

	if len(bytes) == 2 {
		return order.Uint16(bytes[0:])
	} else if len(bytes) == 4 {
		return order.Uint32(bytes[0:])
	} else if len(bytes) == 8 {
		return order.Uint64(bytes[0:])
	} else {
		log.Println("invalid length in ValueOrderChange")
	}
	return 0
}
func HostToNet16(v uint16) uint16 {
	return ValueOrderChange(v, binary.BigEndian).(uint16)
}
func HostToNet32(v uint32) uint32 {
	return ValueOrderChange(v, binary.BigEndian).(uint32)
}
func NetToHost16(v uint16) uint16 {
	return ValueOrderChange(v, binary.LittleEndian).(uint16)
}
func NetToHost32(v uint32) uint32 {
	return ValueOrderChange(v, binary.LittleEndian).(uint32)
}

// ReadBig read a uint16/uint32/uint64(BigEndian) from io.Reader
func ReadBig(r io.Reader, data interface{}) error {
	return binary.Read(r, binary.BigEndian, data)
}

// ReadLittle read a uint16/uint32/uint64(LittleEndian) from io.Reader
func ReadLittle(r io.Reader, data interface{}) error {
	return binary.Read(r, binary.LittleEndian, data)
}

// WriteBig write a uint16/uint32/uint64(BigEndian) to io.Writer
func WriteBig(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.BigEndian, data)
}

// WriteLittle write a uint16/uint32/uint64(LittleEndian) to io.Writer
func WriteLittle(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.LittleEndian, data)
}

// ByteToInt16Slice converts []byte to []int16(LittleEndian).
func ByteToInt16Slice(buf []byte) ([]int16, error) {
	if len(buf)%2 != 0 {
		return nil, errors.New("trailing bytes")
	}
	vals := make([]int16, len(buf)/2)
	for i := 0; i < len(vals); i++ {
		val := binary.LittleEndian.Uint16(buf[i*2:])
		vals[i] = int16(val)
	}
	return vals, nil
}

// Int16ToByteSlice converts []int16(LittleEndian) to []byte.
func Int16ToByteSlice(vals []int16) []byte {
	buf := make([]byte, len(vals)*2)
	for i, v := range vals {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(v))
	}
	return buf
}

// StringPair like std::pair
type StringPair struct {
	First  string
	Second string
}

func (s StringPair) ToString(sp string) string {
	return s.First + sp + s.Second
}

type Uint32Slice []uint32

func (p Uint32Slice) Len() int {
	return len(p)
}

func (p Uint32Slice) Less(i, j int) bool {
	return p[i] <= p[j]
}

func (p Uint32Slice) Swap(i, j int) {
	tmp := p[i]
	p[i] = p[j]
	p[j] = tmp
}

func (p Uint32Slice) Search(x uint32) int {
	for idx, val := range p {
		if val >= x {
			return idx
		}
	}
	return -1
}

// Auto Counter
type AutoCounter int

func (c *AutoCounter) Pre() int {
	val := *c
	*c += 1
	return int(val)
}

func (c *AutoCounter) Post() int {
	*c += 1
	val := *c
	return int(val)
}

// Return copy(capsize) of src and the size of data
func Clone2(src []byte, capsize int) ([]byte, int) {
	dst := make([]byte, capsize)
	size := Min(len(src), capsize)
	copy(dst, src[:size])
	return dst, size
}

// Return copy of src
func Clone(src []byte) []byte {
	capsize := len(src)
	dst, _ := Clone2(src, capsize)
	return dst
}

// Convert []byte to int16[]
func Convert2Int16(src []byte) []int16 {
	src_ptr := C.uint8ptr(unsafe.Pointer(&src[0]))
	src_len := C.size_t(len(src))

	dst := make([]int16, (len(src)+1)/2)
	dst_ptr := C.int16ptr(unsafe.Pointer(&dst[0]))
	dst_len := C.size_t(len(dst))
	C.convert_to_int16(src_ptr, src_len, dst_ptr, dst_len)
	return dst
}

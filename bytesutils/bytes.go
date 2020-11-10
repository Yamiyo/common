package bytesutils

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"

	"github.com/Yamiyo/common/convertutils"
)

// ToHex returns the hex representation of b, prefixed with '0x'.
// For empty slices, the return value is "0x0".
//
// Deprecated: use hexutil.Encode instead.
func ToHex(b []byte) string {
	hex := Bytes2Hex(b)
	if len(hex) == 0 {
		hex = "0"
	}
	return "0x" + hex
}

// ToHexArray creates a array of hex-string based on []byte
func ToHexArray(b [][]byte) []string {
	r := make([]string, len(b))
	for i := range b {
		r[i] = ToHex(b[i])
	}
	return r
}

// FromHex returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func FromHex(s string) []byte {
	if has0xPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// FromInt64ToByteBE ...
func FromInt64ToByteBE(i int64) (b []byte) {
	b = make([]byte, 8)

	b[7] = uint8(i)
	b[6] = uint8(i >> 8)
	b[5] = uint8(i >> 16)
	b[4] = uint8(i >> 24)
	b[3] = uint8(i >> 32)
	b[2] = uint8(i >> 40)
	b[1] = uint8(i >> 48)
	b[0] = uint8(i >> 56)

	return
}

// ToInt64FromByteBE ...
func ToInt64FromByteBE(b []byte) (i int64) {
	i = int64(b[7]) | int64(b[6])<<8 | int64(b[5])<<16 | int64(b[4])<<24 | int64(b[3])<<32 | int64(b[2])<<40 | int64(b[1])<<48 | int64(b[0])<<56
	return
}

// CopyBytes returns an exact copy of the provided bytes.
func CopyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}

// has0xPrefix validates str begins with '0x' or '0X'.
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

// Bytes2Hex returns the hexadecimal encoding of d.
func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// Hex2BytesFixed returns bytes of a specified fixed length flen.
func Hex2BytesFixed(str string, flen int) []byte {
	h, _ := hex.DecodeString(str)
	if len(h) == flen {
		return h
	}
	if len(h) > flen {
		return h[len(h)-flen:]
	}
	hh := make([]byte, flen)
	copy(hh[flen-len(h):flen], h)
	return hh
}

// RightPadBytes zero-pads slice to the right up to length l.
func RightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded, slice)

	return padded
}

// LeftPadBytes zero-pads slice to the left up to length l.
func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

// TrimLeftZeroes returns a subslice of s without leading zeroes
func TrimLeftZeroes(s []byte) []byte {
	idx := 0
	for ; idx < len(s); idx++ {
		if s[idx] != 0 {
			break
		}
	}
	return s[idx:]
}

// BytesCompare ...
func BytesCompare(a, b []byte) int {
	return bytes.Compare(a, b)
}

// BytesEqual ...
func BytesEqual(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// BytesNotEqual ...
func BytesNotEqual(a, b []byte) bool {
	return !bytes.Equal(a, b)
}

// Separator ...
func Separator(dst, a, b []byte) []byte {
	i, n := 0, len(a)
	if n > len(b) {
		n = len(b)
	}
	for ; i < n && a[i] == b[i]; i++ {
	}
	if i >= n {
		// Do not shorten if one string is a prefix of the other
	} else if c := a[i]; c < 0xff && c+1 < b[i] {
		dst = append(dst, a[:i+1]...)
		dst[len(dst)-1]++
		return dst
	}
	return nil
}

// Successor ...
func Successor(dst, b []byte) []byte {
	for i, c := range b {
		if c != 0xff {
			dst = append(dst, b[:i+1]...)
			dst[len(dst)-1]++
			return dst
		}
	}
	return nil
}

// BytesSeparator ...
func BytesSeparator(a, b []byte) []byte {
	if bytes.Equal(a, b) {
		return b
	}
	i, n := 0, len(a)
	if n > len(b) {
		n = len(b)
	}
	for ; i < n && (a[i] == b[i]); i++ {
	}
	x := append([]byte{}, a[:i]...)
	if i < n {
		if c := a[i] + 1; c < b[i] {
			return append(x, c)
		}
		x = append(x, a[i])
		i++
	}
	for ; i < len(a); i++ {
		c := a[i]
		if c < 0xff {
			return append(x, c+1)
		}
		x = append(x, c)
	}
	if len(b) > i && b[i] > 0 {
		return append(x, b[i]-1)
	}
	return append(x, 'x')
}

// BytesAfter ...
func BytesAfter(b []byte) []byte {
	var x []byte
	for _, c := range b {
		if c < 0xff {
			return append(x, c+1)
		}
		x = append(x, c)
	}
	return append(x, 'x')
}

// Sha1 ...
func Sha1(obj []interface{}) (string, error) {
	buf := make([]byte, 0)

	for _, o := range obj {
		if err := convertutils.ConvertToBytes(&buf, o); err != nil {
			return "", err
		}
	}

	sha1 := sha1.Sum(buf)

	return ToHex(sha1[:]), nil
}

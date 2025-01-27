package cloudian

import (
	"fmt"
)

// Construct by e.g. 3 * TB
type ByteSize uint64

const (
	KB ByteSize = 1
	MB          = KB << 10
	GB          = MB << 10
	TB          = GB << 10
)

func (b ByteSize) KB() uint64 {
	return uint64(b)
}

func (b ByteSize) MB() float64 {
	v := b / MB
	r := b % MB
	return float64(v) + float64(r)/float64(MB)
}

func (b ByteSize) GB() float64 {
	v := b / GB
	r := b % GB
	return float64(v) + float64(r)/float64(GB)
}

func (b ByteSize) TB() float64 {
	v := b / TB
	r := b % TB
	return float64(v) + float64(r)/float64(TB)
}

func (b ByteSize) KBString() string {
	if b == 0 {
		return "0"
	}
	return fmt.Sprintf("%d", b/KB)
}

package types

import (
	"fmt"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type DiskStatus struct {
	All  uint64 `json:"All"`
	Used uint64 `json:"Used"`
	Free uint64 `json:"Free"`
}

func (d DiskStatus) FreePercent() float64 {
	return float64(d.Free) / float64(d.All) * 100
}

func (d DiskStatus) FreeOfAll() string {
	str := ""

	str += fmt.Sprintf("%.2F", float64(d.Free)/float64(GB))
	str += "/"
	str += fmt.Sprintf("%.2F", float64(d.All)/float64(GB))
	str += " GB"

	return str
}

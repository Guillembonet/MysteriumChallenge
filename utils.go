package main

import (
	"net"
)

// Exists takes a slice and looks for an element in it. If found it will
// return true else it will return false.
func Exists(slice []net.Addr, val net.Addr) bool {
	for _, item := range slice {
		if item.String() == val.String() {
			return true
		}
	}
	return false
}

// +build !embed

package gorocksdb

// #cgo LDFLAGS: -L/home/jelte/go/src/github.com/GetStream/gorocksdb/lib -L/home/jelte/go/src/github.com/GetStream/Keevo/.rocksdb-repo -lrocksdb -lhello -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -ldl
// #cgo CFLAGS: -I/home/jelte/go/src/github.com/GetStream/Keevo/.rocksdb-repo/include
import "C"

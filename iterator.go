package gorocksdb

// #include <stdlib.h>
// #include "rocksdb/c.h"
// #include "gorocksdb.h"
// #include "lib/hello.h"
import "C"
import (
	"bytes"
	"errors"
	"reflect"
	"unsafe"
)

// Iterator provides a way to seek to specific keys and iterate through
// the keyspace from that point, as well as access the values of those keys.
//
// For example:
//
//      it := db.NewIterator(readOpts)
//      defer it.Close()
//
//      it.Seek([]byte("foo"))
//		for ; it.Valid(); it.Next() {
//          fmt.Printf("Key: %v Value: %v\n", it.Key().Data(), it.Value().Data())
// 		}
//
//      if err := it.Err(); err != nil {
//          return err
//      }
//
type Iterator struct {
	c *C.rocksdb_iterator_t
}

// NewNativeIterator creates a Iterator object.
func NewNativeIterator(c unsafe.Pointer) *Iterator {
	return &Iterator{(*C.rocksdb_iterator_t)(c)}
}

// Valid returns false only when an Iterator has iterated past either the
// first or the last key in the database.
func (iter *Iterator) Valid() bool {
	return C.rocksdb_iter_valid(iter.c) != 0
}

// ValidForPrefix returns false only when an Iterator has iterated past the
// first or the last key in the database or the specified prefix.
func (iter *Iterator) ValidForPrefix(prefix []byte) bool {
	if C.rocksdb_iter_valid(iter.c) == 0 {
		return false
	}

	key := iter.Key()
	result := bytes.HasPrefix(key.Data(), prefix)
	key.Free()
	return result
}

// Key returns the key the iterator currently holds.
func (iter *Iterator) Key() *Slice {
	var cLen C.size_t
	cKey := C.rocksdb_iter_key(iter.c, &cLen)
	if cKey == nil {
		return nil
	}
	return &Slice{cKey, cLen, true}
}

// Value returns the value in the database the iterator currently holds.
func (iter *Iterator) Value() *Slice {
	var cLen C.size_t
	cVal := C.rocksdb_iter_value(iter.c, &cLen)
	if cVal == nil {
		return nil
	}
	return &Slice{cVal, cLen, true}
}

// Next moves the iterator to the next sequential key in the database.
func (iter *Iterator) Next() {
	C.rocksdb_iter_next(iter.c)
}

var ManyKeysPageAllocSize int = 512

type ManyKeys struct {
	c    *C.gorocksdb_many_keys_t
	rust *C.rust_rocksdb_many_keys_t
}

func (m *ManyKeys) Destroy() {
	if m.c != nil {
		C.gorocksdb_destroy_many_keys(m.c)
	}
	if m.rust != nil {
		//C.rust_rocksdb_destroy_many_keys(m.c)
	}
}

func (m *ManyKeys) Found() int {
	if m.c != nil {
		return int(m.c.found)
	}
	return int(m.rust.found)
}

func (m *ManyKeys) Keys() [][]byte {
	found := m.Found()
	var cKeys **C.char
	var cKeySizes *C.size_t
	if m.c != nil {
		cKeys = m.c.keys
		cKeySizes = m.c.key_sizes
	} else {
		cKeys = m.rust.keys
		cKeySizes = m.rust.key_sizes
	}

	keys := make([][]byte, found)

	for i := uintptr(0); i < uintptr(found); i++ {
		chars := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cKeys)) + i*unsafe.Sizeof(cKeys)))
		size := *(*C.size_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cKeySizes)) + i*unsafe.Sizeof(cKeySizes)))
		keys[i] = charToByte(chars, size)

	}
	return keys
}

func (m *ManyKeys) Values() [][]byte {
	found := m.Found()
	values := make([][]byte, found)
	var cValues **C.char
	var cValueSizes *C.size_t
	if m.c != nil {
		cValues = m.c.values
		cValueSizes = m.c.value_sizes
	} else {
		cValues = m.rust.values
		cValueSizes = m.rust.value_sizes
	}

	for i := uintptr(0); i < uintptr(found); i++ {
		chars := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cValues)) + i*unsafe.Sizeof(cValues)))
		size := *(*C.size_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cValueSizes)) + i*unsafe.Sizeof(cValueSizes)))
		values[i] = charToByte(chars, size)
	}
	return values
}

func (m *ManyKeys) Each(each func(i int, key []byte, value []byte) bool) bool {
	found := m.Found()
	for i := uintptr(0); i < uintptr(found); i++ {
		chars := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(m.c.keys)) + i*unsafe.Sizeof(m.c.keys)))
		size := *(*C.size_t)(unsafe.Pointer(uintptr(unsafe.Pointer(m.c.key_sizes)) + i*unsafe.Sizeof(m.c.key_sizes)))
		key := charToByte(chars, size)

		chars = *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(m.c.values)) + i*unsafe.Sizeof(m.c.values)))
		size = *(*C.size_t)(unsafe.Pointer(uintptr(unsafe.Pointer(m.c.value_sizes)) + i*unsafe.Sizeof(m.c.value_sizes)))
		value := charToByte(chars, size)

		if !each(int(i), key, value) {
			return false
		}
	}
	return true
}

//....
func (iter *Iterator) NextManyKeys(size int) *ManyKeys {
	return &ManyKeys{c: C.gorocksdb_iter_next_many_keys(iter.c, C.int(size))}
}

//....
func (iter *Iterator) NextManyKeysF(limit int, keyPrefix, keyEnd []byte) *ManyKeys {
	cKeyFilter := C.gorocksdb_many_keys_filter_t{}
	if len(keyPrefix) > 0 {
		cKeyPrefix := C.CString(string(keyPrefix))
		defer C.free(unsafe.Pointer(cKeyPrefix))
		cKeyFilter.key_prefix = cKeyPrefix
		cKeyFilter.key_prefix_s = C.size_t(len(keyPrefix))
	}
	if len(keyEnd) > 0 {
		cKeyEnd := C.CString(string(keyEnd))
		defer C.free(unsafe.Pointer(cKeyEnd))
		cKeyFilter.key_end = cKeyEnd
		cKeyFilter.key_end_s = C.size_t(len(keyEnd))
	}
	return &ManyKeys{c: C.gorocksdb_iter_next_many_keys_f(iter.c, C.int(limit), &cKeyFilter, C.int(ManyKeysPageAllocSize))}
}

func (iter *Iterator) RustNextManyKeysF(limit int, keyPrefix, keyStop []byte) *ManyKeys {
	result := C.rust_rocksdb_iter_next_many_keys_f(
		iter.c,
		C.int(limit),
		C.rust_rocksdb_many_keys_filter_t{
			key_prefix:      byteToChar(keyPrefix),
			key_prefix_size: C.size_t(len(keyPrefix)),
			key_stop:        byteToChar(keyStop),
			key_stop_size:   C.size_t(len(keyStop)),
		},
	)
	return &ManyKeys{rust: &result}
}

type KeysSearch struct {
	KeyFrom, KeyPrefix, KeyEnd []byte
	Limit                      int
}

type ManyManyKeys struct {
	c    **C.gorocksdb_many_keys_t
	size int
}

func (iter *Iterator) ManySearchKeys(searches []KeysSearch) *ManyManyKeys {
	nbSearches := len(searches)
	cManyKeysSearches := make([]C.gorocksdb_keys_search_t, nbSearches)
	for i := range searches {
		cKSearch := C.gorocksdb_keys_search_t{limit: C.int(searches[i].Limit)}
		cKSearch.key_from = C.CString(string(searches[i].KeyFrom))
		cKSearch.key_from_s = C.size_t(len(searches[i].KeyFrom))
		if len(searches[i].KeyPrefix) > 0 {
			cKSearch.key_prefix = C.CString(string(searches[i].KeyPrefix))
			cKSearch.key_prefix_s = C.size_t(len(searches[i].KeyPrefix))
		}
		if len(searches[i].KeyEnd) > 0 {
			cKSearch.key_end = C.CString(string(searches[i].KeyEnd))
			cKSearch.key_end_s = C.size_t(len(searches[i].KeyEnd))
		}
		cManyKeysSearches[i] = cKSearch
	}
	cManyManyKeys := C.gorocksdb_many_search_keys(iter.c,
		(*C.gorocksdb_keys_search_t)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&cManyKeysSearches)).Data)),
		C.int(nbSearches),
		C.int(ManyKeysPageAllocSize),
	)
	// free
	for i := range searches {
		C.free(unsafe.Pointer(cManyKeysSearches[i].key_from))
		if len(searches[i].KeyPrefix) > 0 {
			C.free(unsafe.Pointer(cManyKeysSearches[i].key_prefix))
		}
		if len(searches[i].KeyEnd) > 0 {
			C.free(unsafe.Pointer(cManyKeysSearches[i].key_end))
		}
	}
	return &ManyManyKeys{c: cManyManyKeys, size: nbSearches}
}

func (m ManyManyKeys) Result() []*ManyKeys {
	result := make([]*ManyKeys, m.size)
	for i := uintptr(0); i < uintptr(m.size); i++ {
		manyKeys := *(**C.gorocksdb_many_keys_t)(unsafe.Pointer(uintptr(unsafe.Pointer(m.c)) + i*unsafe.Sizeof(m.c)))
		result[i] = &ManyKeys{c: manyKeys}
	}
	return result
}

func (m ManyManyKeys) Destroy() {
	C.gorocksdb_destroy_many_many_keys(m.c, C.int(m.size))
}

// Prev moves the iterator to the previous sequential key in the database.
func (iter *Iterator) Prev() {
	C.rocksdb_iter_prev(iter.c)
}

// SeekToFirst moves the iterator to the first key in the database.
func (iter *Iterator) SeekToFirst() {
	C.rocksdb_iter_seek_to_first(iter.c)
}

// SeekToLast moves the iterator to the last key in the database.
func (iter *Iterator) SeekToLast() {
	C.rocksdb_iter_seek_to_last(iter.c)
}

// Seek moves the iterator to the position greater than or equal to the key.
func (iter *Iterator) Seek(key []byte) {
	cKey := byteToChar(key)
	C.rocksdb_iter_seek(iter.c, cKey, C.size_t(len(key)))
}

// SeekForPrev moves the iterator to the last key that less than or equal
// to the target key, in contrast with Seek.
func (iter *Iterator) SeekForPrev(key []byte) {
	cKey := byteToChar(key)
	C.rocksdb_iter_seek_for_prev(iter.c, cKey, C.size_t(len(key)))
}

// Err returns nil if no errors happened during iteration, or the actual
// error otherwise.
func (iter *Iterator) Err() error {
	var cErr *C.char
	C.rocksdb_iter_get_error(iter.c, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// Close closes the iterator.
func (iter *Iterator) Close() {
	C.rocksdb_iter_destroy(iter.c)
	iter.c = nil
}

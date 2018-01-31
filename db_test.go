package gorocksdb

import (
	"io/ioutil"
	"testing"

	"github.com/facebookgo/ensure"
)

func TestOpenDb(t *testing.T) {
	db := newTestDB(t, "TestOpenDb", nil)
	defer db.Close()
}

func TestDBCRUD(t *testing.T) {
	db := newTestDB(t, "TestDBGet", nil)
	defer db.Close()

	var (
		givenKey  = []byte("hello")
		givenVal1 = []byte("world1")
		givenVal2 = []byte("world2")
		wo        = NewDefaultWriteOptions()
		ro        = NewDefaultReadOptions()
	)

	// create
	ensure.Nil(t, db.Put(wo, givenKey, givenVal1))

	// retrieve
	v1, err := db.Get(ro, givenKey)
	defer v1.Free()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, v1.Data(), givenVal1)

	// update
	ensure.Nil(t, db.Put(wo, givenKey, givenVal2))
	v2, err := db.Get(ro, givenKey)
	defer v2.Free()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, v2.Data(), givenVal2)

	// delete
	ensure.Nil(t, db.Delete(wo, givenKey))
	v3, err := db.Get(ro, givenKey)
	ensure.Nil(t, err)
	ensure.True(t, v3.Data() == nil)
}

func newTestDB(t *testing.T, name string, applyOpts func(opts *Options)) *DB {
	dir, err := ioutil.TempDir("", "gorocksdb-"+name)
	ensure.Nil(t, err)

	opts := NewDefaultOptions()
	// test the ratelimiter
	rateLimiter := NewRateLimiter(1024, 100*1000, 10)
	opts.SetRateLimiter(rateLimiter)
	opts.SetCreateIfMissing(true)
	if applyOpts != nil {
		applyOpts(opts)
	}
	db, err := OpenDb(opts, dir)
	ensure.Nil(t, err)

	return db
}

func TestDBMultiGet(t *testing.T) {
	db := newTestDB(t, "TestDBMultiGet", nil)
	defer db.Close()

	var (
		givenKey1 = []byte("hello1")
		givenKey2 = []byte("hello2")
		givenKey3 = []byte("hello3")
		givenVal1 = []byte("world1")
		givenVal2 = []byte("world2")
		givenVal3 = []byte("world3")
		wo        = NewDefaultWriteOptions()
		ro        = NewDefaultReadOptions()
	)

	// create
	ensure.Nil(t, db.Put(wo, givenKey1, givenVal1))
	ensure.Nil(t, db.Put(wo, givenKey2, givenVal2))
	ensure.Nil(t, db.Put(wo, givenKey3, givenVal3))

	// retrieve
	values, err := db.MultiGet(ro, []byte("noexist"), givenKey1, givenKey2, givenKey3)
	defer values.Destroy()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(values), 4)

	ensure.DeepEqual(t, values[0].Data(), []byte(nil))
	ensure.DeepEqual(t, values[1].Data(), givenVal1)
	ensure.DeepEqual(t, values[2].Data(), givenVal2)
	ensure.DeepEqual(t, values[3].Data(), givenVal3)
}

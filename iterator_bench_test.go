package gorocksdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type NextManyKeysF func(limit int, keyPrefix []byte, keyEnd []byte) *ManyKeys

type BenchSuite struct {
	suite.BenchmarkSuite
	db            *DB
	NextManyKeysF func(*Iterator) NextManyKeysF
}

// Run is a helper to do a sub benchmark
func (s *BenchSuite) Run(name string, f func()) {
	oldB := s.B()
	s.B().Run(name, func(b *testing.B) {
		s.SetB(b)

		defer func() {
			s.TearDownTest()
			s.SetB(oldB)
		}()

		s.SetupTest()
		f()
	})
}

// N returns the N of the current B()
func (s *BenchSuite) N() int {
	return s.B().N
}

func setMinRuns(b *testing.B) {
	if b.N < 5 {
		b.N = 5
	}
}

func (s *BenchSuite) SetupTest() {
	setMinRuns(s.B())
	s.db = newTestDB(s.B(), "BenchIterator", nil)
	s.B().ResetTimer()
}

func (s *BenchSuite) TearDownTest() {
	s.B().StopTimer()
	opts := NewDefaultOptions()
	defer opts.Destroy()
	s.db.Close()
	s.NoError(DestroyDb(s.db.name, opts))
}

func (s *BenchSuite) BenchmarkNextManyKeysF() {
	nKeys := 1000
	wo := NewDefaultWriteOptions()
	defer wo.Destroy()
	for i := 0; i < nKeys; i++ {
		k := fmt.Sprintf("A/%d", i)
		s.NoError(s.db.Put(wo, []byte(k), []byte("val_"+k)))
	}
	ro := NewDefaultReadOptions()
	defer ro.Destroy()
	// s.iter = s.db.NewIterator(ro)
	// defer s.iter.Close()

	s.B().ResetTimer()

	for n := 0; n < s.N(); n++ {
		iter := s.db.NewIterator(ro)
		iter.SeekToFirst()
		manyKeys := s.NextManyKeysF(iter)(nKeys, []byte(""), nil)
		manyKeys.Destroy()
		iter.Close()
	}

	s.B().StopTimer()
}

func BenchmarkC(b *testing.B) {
	suite.RunBenchmark(b, &BenchSuite{NextManyKeysF: func(iter *Iterator) NextManyKeysF {
		return NextManyKeysF(iter.NextManyKeysF)
	}})
}

func BenchmarkRust(b *testing.B) {
	suite.RunBenchmark(b, &BenchSuite{NextManyKeysF: func(iter *Iterator) NextManyKeysF {
		return NextManyKeysF(iter.RustNextManyKeysF)
	}})
}

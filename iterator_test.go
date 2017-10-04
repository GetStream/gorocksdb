package gorocksdb

import (
	"testing"

	"github.com/facebookgo/ensure"
)

func TestIterator(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val")))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()
	var actualKeys [][]byte
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		key := make([]byte, 4)
		copy(key, iter.Key().Data())
		actualKeys = append(actualKeys, key)
	}
	ensure.Nil(t, iter.Err())
	ensure.DeepEqual(t, actualKeys, givenKeys)
}

func TestIteratorMany(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val")))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()
	var actualKeys [][]byte
	iter.SeekToFirst()

	manyKeys := iter.NextManyKeys(2)
	for manyKeys.Found() > 0 {
		for _, k := range manyKeys.Keys() {
			newK := make([]byte, len(k))
			copy(newK, k)
			actualKeys = append(actualKeys, newK)
		}
		manyKeys.Destroy()
		manyKeys = iter.NextManyKeys(2)
	}
	manyKeys.Destroy()
	ensure.Nil(t, iter.Err())
	ensure.DeepEqual(t, actualKeys, givenKeys)
}

func TestIteratorManyFOnKeyPrefix(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3"), []byte("keyB1")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()
	var actualKeys [][]byte
	var actualValues [][]byte
	iter.SeekToFirst()

	manyKeys := iter.NextManyKeysF(2, []byte("keyA"), nil)
	for manyKeys.Found() > 0 {
		for _, k := range manyKeys.Keys() {
			newK := make([]byte, len(k))
			copy(newK, k)
			actualKeys = append(actualKeys, newK)
		}
		for _, v := range manyKeys.Values() {
			newV := make([]byte, len(v))
			copy(newV, v)
			actualValues = append(actualValues, newV)
		}
		manyKeys.Destroy()
		manyKeys = iter.NextManyKeysF(2, []byte("keyA"), nil)
	}
	manyKeys.Destroy()
	ensure.Nil(t, iter.Err())
	ensure.DeepEqual(t, actualKeys, [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3")})
	ensure.DeepEqual(t, actualValues, [][]byte{[]byte("val_keyA1"), []byte("val_keyA2"), []byte("val_keyA3")})
}

func TestIteratorManyFOnKeyPrefixRust(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3"), []byte("keyB1")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()
	var actualKeys [][]byte
	var actualValues [][]byte
	iter.SeekToFirst()

	manyKeys := iter.RustNextManyKeysF(2, []byte("keyA"), nil)
	for manyKeys.Found() > 0 {
		for _, k := range manyKeys.Keys() {
			newK := make([]byte, len(k))
			copy(newK, k)
			actualKeys = append(actualKeys, newK)
		}
		for _, v := range manyKeys.Values() {
			newV := make([]byte, len(v))
			copy(newV, v)
			actualValues = append(actualValues, newV)
		}
		manyKeys.Destroy()
		manyKeys = iter.NextManyKeysF(2, []byte("keyA"), nil)
	}
	manyKeys.Destroy()
	ensure.Nil(t, iter.Err())
	ensure.DeepEqual(t, actualKeys, [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3")})
	ensure.DeepEqual(t, actualValues, [][]byte{[]byte("val_keyA1"), []byte("val_keyA2"), []byte("val_keyA3")})
}

func TestIteratorManyFWithLimit(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3"), []byte("keyA4")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()

	iter.SeekToFirst()
	manyKeys := iter.NextManyKeysF(-1, []byte("keyA"), nil)
	ensure.DeepEqual(t, manyKeys.Keys(), [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3"), []byte("keyA4")})
	ensure.DeepEqual(t, manyKeys.Values(), [][]byte{[]byte("val_keyA1"), []byte("val_keyA2"), []byte("val_keyA3"), []byte("val_keyA4")})
	manyKeys.Destroy()

	iter.SeekToFirst()
	manyKeys = iter.NextManyKeysF(0, []byte("keyA"), nil)
	ensure.DeepEqual(t, manyKeys.Keys(), [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3"), []byte("keyA4")})
	ensure.DeepEqual(t, manyKeys.Values(), [][]byte{[]byte("val_keyA1"), []byte("val_keyA2"), []byte("val_keyA3"), []byte("val_keyA4")})
	manyKeys.Destroy()

	iter.SeekToFirst()
	manyKeys = iter.NextManyKeysF(2, []byte("keyA"), nil)
	ensure.DeepEqual(t, manyKeys.Keys(), [][]byte{[]byte("keyA1"), []byte("keyA2")})
	ensure.DeepEqual(t, manyKeys.Values(), [][]byte{[]byte("val_keyA1"), []byte("val_keyA2")})
	manyKeys.Destroy()

	iter.SeekToFirst()
	manyKeys = iter.NextManyKeysF(20, []byte("keyA"), nil)
	ensure.DeepEqual(t, manyKeys.Keys(), [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3"), []byte("keyA4")})
	ensure.DeepEqual(t, manyKeys.Values(), [][]byte{[]byte("val_keyA1"), []byte("val_keyA2"), []byte("val_keyA3"), []byte("val_keyA4")})
	manyKeys.Destroy()
}

func TestIteratorManyFOnKeyEnd(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("A"), []byte("B"), []byte("C"), []byte("C1"), []byte("D")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()
	var actualKeys [][]byte
	var actualValues [][]byte
	iter.SeekToFirst()

	manyKeys := iter.NextManyKeysF(2, nil, []byte("C1"))
	for manyKeys.Found() > 0 {
		for _, k := range manyKeys.Keys() {
			newK := make([]byte, len(k))
			copy(newK, k)
			actualKeys = append(actualKeys, newK)
		}
		for _, v := range manyKeys.Values() {
			newV := make([]byte, len(v))
			copy(newV, v)
			actualValues = append(actualValues, newV)
		}
		manyKeys.Destroy()
		manyKeys = iter.NextManyKeysF(2, nil, []byte("C1"))
	}
	manyKeys.Destroy()
	ensure.Nil(t, iter.Err())
	ensure.DeepEqual(t, actualKeys, [][]byte{[]byte("A"), []byte("B"), []byte("C")})
	ensure.DeepEqual(t, actualValues, [][]byte{[]byte("val_A"), []byte("val_B"), []byte("val_C")})
}

func TestIteratorManyFOnKeyPrefixAndEnd(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("keyA"), []byte("keyB"), []byte("keyC"), []byte("keyC1")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()
	var actualKeys [][]byte
	var actualValues [][]byte
	iter.SeekToFirst()

	manyKeys := iter.NextManyKeysF(2, []byte("key"), []byte("keyC1"))
	for manyKeys.Found() > 0 {
		for _, k := range manyKeys.Keys() {
			newK := make([]byte, len(k))
			copy(newK, k)
			actualKeys = append(actualKeys, newK)
		}
		for _, v := range manyKeys.Values() {
			newV := make([]byte, len(v))
			copy(newV, v)
			actualValues = append(actualValues, newV)
		}
		manyKeys.Destroy()
		manyKeys = iter.NextManyKeysF(2, []byte("key"), []byte("keyC1"))
	}
	manyKeys.Destroy()
	ensure.Nil(t, iter.Err())
	ensure.DeepEqual(t, actualKeys, [][]byte{[]byte("keyA"), []byte("keyB"), []byte("keyC")})
	ensure.DeepEqual(t, actualValues, [][]byte{[]byte("val_keyA"), []byte("val_keyB"), []byte("val_keyC")})
}

func TestIteratorManySearchKeys(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("A"), []byte("B"), []byte("C"), []byte("D"), []byte("E"), []byte("F")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()

	searches := make([]KeysSearch, 3)
	searches[0] = KeysSearch{KeyFrom: []byte("A"), Limit: 1000}
	searches[1] = KeysSearch{KeyFrom: []byte("D"), Limit: 1000}
	searches[2] = KeysSearch{KeyFrom: []byte("Z"), Limit: 1000}

	manyManyKeys := iter.ManySearchKeys(searches)
	defer manyManyKeys.Destroy()
	result := manyManyKeys.Result()
	if len(result) != len(searches) {
		t.Fatalf("result len should be %d", len(searches))
	}
	ensure.DeepEqual(t, result[0].Found(), 6)
	ensure.DeepEqual(t, result[0].Keys(), [][]byte{[]byte("A"), []byte("B"), []byte("C"), []byte("D"), []byte("E"), []byte("F")})
	ensure.DeepEqual(t, result[0].Values(), [][]byte{[]byte("val_A"), []byte("val_B"), []byte("val_C"), []byte("val_D"), []byte("val_E"), []byte("val_F")})
	ensure.DeepEqual(t, result[1].Found(), 3)
	ensure.DeepEqual(t, result[1].Keys(), [][]byte{[]byte("D"), []byte("E"), []byte("F")})
	ensure.DeepEqual(t, result[1].Values(), [][]byte{[]byte("val_D"), []byte("val_E"), []byte("val_F")})
	ensure.DeepEqual(t, result[2].Found(), 0)
	ensure.DeepEqual(t, result[2].Keys(), [][]byte{})
	ensure.DeepEqual(t, result[2].Values(), [][]byte{})
}

func TestIteratorManySearchKeysWithKeyPrefix(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("A1"), []byte("A2"), []byte("B1"), []byte("C1"), []byte("D1"), []byte("D2")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()

	searches := make([]KeysSearch, 4)
	searches[0] = KeysSearch{KeyFrom: []byte("A"), KeyPrefix: []byte("A"), Limit: 1000}
	searches[1] = KeysSearch{KeyFrom: []byte("B"), KeyPrefix: []byte("B"), Limit: 1000}
	searches[2] = KeysSearch{KeyFrom: []byte("D"), KeyPrefix: []byte("D"), Limit: 1000}
	searches[3] = KeysSearch{KeyFrom: []byte("Z"), KeyPrefix: []byte("Z"), Limit: 1000}

	manyManyKeys := iter.ManySearchKeys(searches)
	defer manyManyKeys.Destroy()
	result := manyManyKeys.Result()
	if len(result) != len(searches) {
		t.Fatalf("result len should be %d", len(searches))
	}
	ensure.DeepEqual(t, result[0].Found(), 2)
	ensure.DeepEqual(t, result[0].Keys(), [][]byte{[]byte("A1"), []byte("A2")})
	ensure.DeepEqual(t, result[0].Values(), [][]byte{[]byte("val_A1"), []byte("val_A2")})
	ensure.DeepEqual(t, result[1].Found(), 1)
	ensure.DeepEqual(t, result[1].Keys(), [][]byte{[]byte("B1")})
	ensure.DeepEqual(t, result[1].Values(), [][]byte{[]byte("val_B1")})
	ensure.DeepEqual(t, result[2].Found(), 2)
	ensure.DeepEqual(t, result[2].Keys(), [][]byte{[]byte("D1"), []byte("D2")})
	ensure.DeepEqual(t, result[2].Values(), [][]byte{[]byte("val_D1"), []byte("val_D2")})
	ensure.DeepEqual(t, result[3].Found(), 0)
	ensure.DeepEqual(t, result[3].Keys(), [][]byte{})
	ensure.DeepEqual(t, result[3].Values(), [][]byte{})
}

func TestIteratorManySearchKeysWithKeyEnd(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("A1"), []byte("A2"), []byte("A3"), []byte("B1"), []byte("B2"), []byte("B3")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()

	searches := make([]KeysSearch, 2)
	searches[0] = KeysSearch{KeyFrom: []byte("A"), KeyEnd: []byte("A3"), Limit: 1000}
	searches[1] = KeysSearch{KeyFrom: []byte("B"), KeyEnd: []byte("B2"), Limit: 1000}

	manyManyKeys := iter.ManySearchKeys(searches)
	defer manyManyKeys.Destroy()
	result := manyManyKeys.Result()
	if len(result) != len(searches) {
		t.Fatalf("result len should be %d", len(searches))
	}
	ensure.DeepEqual(t, result[0].Found(), 2)
	ensure.DeepEqual(t, result[0].Keys(), [][]byte{[]byte("A1"), []byte("A2")})
	ensure.DeepEqual(t, result[0].Values(), [][]byte{[]byte("val_A1"), []byte("val_A2")})
	ensure.DeepEqual(t, result[1].Found(), 1)
	ensure.DeepEqual(t, result[1].Keys(), [][]byte{[]byte("B1")})
	ensure.DeepEqual(t, result[1].Values(), [][]byte{[]byte("val_B1")})
}

func TestIteratorManySearchKeysWithKeyPrefixAndEnd(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("keyC"), []byte("keyC0"), []byte("keyC1")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()

	searches := make([]KeysSearch, 2)
	searches[0] = KeysSearch{KeyFrom: []byte("keyC0"), KeyPrefix: []byte("keyC"), KeyEnd: []byte("keyC1"), Limit: 1000}
	searches[1] = KeysSearch{KeyFrom: []byte("k"), KeyPrefix: []byte("keyC"), KeyEnd: []byte("keyC1"), Limit: 1000}

	manyManyKeys := iter.ManySearchKeys(searches)
	defer manyManyKeys.Destroy()
	result := manyManyKeys.Result()
	if len(result) != len(searches) {
		t.Fatalf("result len should be %d", len(searches))
	}
	ensure.DeepEqual(t, result[0].Found(), 1)
	ensure.DeepEqual(t, result[0].Keys(), [][]byte{[]byte("keyC0")})
	ensure.DeepEqual(t, result[0].Values(), [][]byte{[]byte("val_keyC0")})
	ensure.DeepEqual(t, result[1].Found(), 2)
	ensure.DeepEqual(t, result[1].Keys(), [][]byte{[]byte("keyC"), []byte("keyC0")})
	ensure.DeepEqual(t, result[1].Values(), [][]byte{[]byte("val_keyC"), []byte("val_keyC0")})
}

func TestIteratorManyKeysEach(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3"), []byte("keyA4")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val_"+string(k))))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()

	iter.SeekToFirst()
	manyKeys := iter.NextManyKeysF(-1, []byte("keyA"), nil)

	actualKeys := [][]byte{}
	actualValues := [][]byte{}
	all := manyKeys.Each(func(i int, key []byte, value []byte) bool {
		actualKeys = append(actualKeys, key)
		actualValues = append(actualValues, value)
		return true
	})
	ensure.DeepEqual(t, all, true)
	ensure.DeepEqual(t, actualKeys, [][]byte{[]byte("keyA1"), []byte("keyA2"), []byte("keyA3"), []byte("keyA4")})
	ensure.DeepEqual(t, actualValues, [][]byte{[]byte("val_keyA1"), []byte("val_keyA2"), []byte("val_keyA3"), []byte("val_keyA4")})

	actualKeys = nil
	actualValues = nil
	limit := 2
	all = manyKeys.Each(func(i int, key []byte, value []byte) bool {
		actualKeys = append(actualKeys, key)
		actualValues = append(actualValues, value)
		return len(actualKeys) != limit
	})
	ensure.DeepEqual(t, all, false)
	ensure.DeepEqual(t, actualKeys, [][]byte{[]byte("keyA1"), []byte("keyA2")})
	ensure.DeepEqual(t, actualValues, [][]byte{[]byte("val_keyA1"), []byte("val_keyA2")})

	manyKeys.Destroy()
}
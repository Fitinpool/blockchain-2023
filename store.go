package main

import (
	"encoding/json"
	"fmt"
	"sync"

	e "blockchain/entities"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Store struct {
	// db *leveldb.DB
	db DataStore
	mu sync.RWMutex
}

func NewStore(dbName string) (*Store, error) {
	pathToDB := fmt.Sprintf("data/%s", dbName)
	db, err := leveldb.OpenFile(pathToDB, nil)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func CopyStore(oldDbName string, newDbName string) (*Store, error) {
	oldPathToDB := fmt.Sprintf("data/%s", oldDbName)
	oldDb, err := leveldb.OpenFile(oldPathToDB, nil)
	if err != nil {
		return nil, err
	}
	defer oldDb.Close()

	newPathToDB := fmt.Sprintf("data/%s", newDbName)
	newDb, err := leveldb.OpenFile(newPathToDB, nil)
	if err != nil {
		return nil, err
	}

	iter := oldDb.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		err = newDb.Put(key, value, nil)
		if err != nil {
			return nil, errors.Wrap(err, "CopyStore newDb.Put error")
		}
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return nil, err
	}

	return &Store{db: newDb}, nil
}

func (st *Store) Put(key string, value interface{}) error {
	st.mu.Lock()
	inputKey := []byte(key)
	inputData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = st.db.Put(inputKey, inputData, nil)
	if err != nil {
		return errors.Wrap(err, "Store.Put st.db.Put error")
	}

	st.mu.Unlock()
	return nil
}

func (st *Store) Get(key string) ([]byte, error) {
	st.mu.Lock()
	inputKey := []byte(key)
	data, err := st.db.Get(inputKey, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Store.Get st.db.Get error")
	}

	st.mu.Unlock()
	return data, nil
}

func (st *Store) Close() error {
	err := st.db.Close()
	if err != nil {
		return errors.Wrap(err, "Store.Close st.db.Close error")
	}

	return nil
}

func (st *Store) IsEmpty() (bool, error) {
	st.mu.Lock()
	iterator := st.db.NewIterator(nil, nil)
	defer iterator.Release()

	if iterator.Next() {
		return false, nil
	}

	if err := iterator.Error(); err != nil {
		return false, errors.Wrap(err, "Store.IsEmpty iterator.Error error")
	}

	st.mu.Unlock()
	return true, nil
}

func (st *Store) GetLastKey() []byte {
	st.mu.Lock()
	iter := st.db.NewIterator(nil, nil)
	var lastValue []byte
	for iter.Next() {
		lastValue = iter.Value()
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		errors.Wrap(err, "Store.GetLastKey iterator.Error error")
	}

	st.mu.Unlock()
	return lastValue
}

func (st *Store) GetAllBlocks() ([]e.Block, error) {
	st.mu.Lock()
	iter := st.db.NewIterator(nil, nil)
	var blocks []e.Block
	for iter.Next() {
		var block e.Block
		err := json.Unmarshal(iter.Value(), &block)
		if err != nil {
			return nil, errors.Wrap(err, "Store.GetAllBlocks json.Unmarshal error")
		}
		blocks = append(blocks, block)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, errors.Wrap(err, "Store.GetAllBlocks iterator.Error error")
	}

	st.mu.Unlock()
	return blocks, nil
}

func (st *Store) GetAllUser() ([]struct {
	Key  string
	Data e.User
}, error) {
	st.mu.Lock()
	iter := st.db.NewIterator(nil, nil)
	var results []struct {
		Key  string
		Data e.User
	}
	for iter.Next() {
		var user e.User
		err := json.Unmarshal(iter.Value(), &user)
		if err != nil {
			return nil, errors.Wrap(err, "Store.GetAllUser json.Unmarshal error")
		}
		results = append(results, struct {
			Key  string
			Data e.User
		}{
			Key:  string(iter.Key()),
			Data: user,
		})
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, errors.Wrap(err, "Store.GetAllUser iterator.Error error")
	}

	st.mu.Unlock()
	return results, nil
}

type DataStore interface {
	Put(key, value []byte, wo *opt.WriteOptions) error
	Get(key []byte, ro *opt.ReadOptions) ([]byte, error)
	Close() error
	NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator
}

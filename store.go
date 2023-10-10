package main

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Store struct {
	// db *leveldb.DB
	db DataStore
}

func NewStore(dbName string) (*Store, error) {
	pathToDB := fmt.Sprintf("data/%s", dbName)
	db, err := leveldb.OpenFile(pathToDB, nil)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (st *Store) Put(key string, value interface{}) error {
	inputKey := []byte(key)
	inputData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = st.db.Put(inputKey, inputData, nil)
	if err != nil {
		return errors.Wrap(err, "Store.Put st.db.Put error")
	}

	return nil
}

func (st *Store) Get(key string) ([]byte, error) {
	inputKey := []byte(key)
	data, err := st.db.Get(inputKey, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Store.Get st.db.Get error")
	}

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
	iterator := st.db.NewIterator(nil, nil)
	defer iterator.Release()

	if iterator.Next() {
		return false, nil
	}

	if err := iterator.Error(); err != nil {
		return false, errors.Wrap(err, "Store.IsEmpty iterator.Error error")
	}

	return true, nil
}

func (st *Store) GetLastKey() []byte {
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
	return lastValue
}

func (st *Store) GetAllBlocks() ([]Block, error) {
	iter := st.db.NewIterator(nil, nil)
	var blocks []Block
	for iter.Next() {
		var block Block
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
	return blocks, nil
}

type DataStore interface {
	Put(key, value []byte, wo *opt.WriteOptions) error
	Get(key []byte, ro *opt.ReadOptions) ([]byte, error)
	Close() error
	NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator
}

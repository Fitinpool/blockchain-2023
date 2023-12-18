package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

type DataItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// func DelStore(dbName string) error {
// 	pathToDB := fmt.Sprintf("data/%s", dbName)
// 	err := os.RemoveAll(pathToDB)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func NewStore(dbName string) (*Store, error) {
	pathToDB := fmt.Sprintf("data/%s", dbName)
	db, err := leveldb.OpenFile(pathToDB, nil)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func ExportarLevelDB(db *Store, archivoDestino string) error {
	iter := db.db.NewIterator(nil, nil)
	defer iter.Release()

	file, err := os.Create(archivoDestino)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		item := DataItem{Key: string(key), Value: string(value)}
		if err := encoder.Encode(item); err != nil {
			return err
		}
	}

	return iter.Error()
}

func ImportarLevelDB(db *Store, archivoOrigen string) error {
	file, err := os.Open(archivoOrigen)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var item DataItem
		if err := decoder.Decode(&item); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := db.db.Put([]byte(item.Key), []byte(item.Value), nil); err != nil {
			return err
		}
	}

	return nil
}

func CopyStore(srcDB, destDB *Store) error {
	if srcDB == nil || destDB == nil {
		return errors.New("srcDB and destDB must not be nil")
	}

	if srcDB.db == nil || destDB.db == nil {
		return errors.New("srcDB.db and destDB.db must not be nil")
	}

	iter := srcDB.db.NewIterator(nil, nil)
	if iter == nil {
		return errors.New("failed to create iterator")
	}

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		// Logging para depuración
		fmt.Printf("Copying key: %s, value: %s\n", key, value)

		if err := destDB.db.Put(key, value, nil); err != nil {
			iter.Release()
			return errors.Wrap(err, "CopyStore destDB.Put error")
		}
	}
	iter.Release()

	return iter.Error()
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

func (st *Store) GetAllBlocks() ([]e.Block, error) {
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

	return blocks, nil
}

func (st *Store) GetAllUser() ([]struct {
	Key  string
	Data e.User
}, error) {
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

	return results, nil
}

type DataStore interface {
	Put(key, value []byte, wo *opt.WriteOptions) error
	Get(key []byte, ro *opt.ReadOptions) ([]byte, error)
	Close() error
	NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator
}

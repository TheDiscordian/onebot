package onelib

import (
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

type levelDB struct {
	path string
	dB   *leveldb.DB
}

func openLevelDB(path string) *levelDB {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		Error.Panicln("Error opening levelDB database:", err)
	}
	return &levelDB{path: path, dB: db}
}

// Get retrieves value by key directly
func (db *levelDB) Get(table, key string) ([]byte, error) {
	data, err := db.dB.Get([]byte(fmt.Sprintf("%s.%s", table, key)), nil)
	return data, err
}

// Searches for key in field, containing key (IE: field:'username', key:'admin'), using an index if exists. Can be very
// slow without an index.
func (db *levelDB) Search(table, field, key string) ([]byte, error) {
	if field == "_id" {
		return db.Get(table, key)
	} else {
		return nil, errors.New("Searching on LevelDB not yet implemented.")
	}
}

// Inserts value into key, erasing any potential previous value.
func (db *levelDB) Put(table, key string, value []byte) error {
	return db.dB.Put([]byte(key), value, nil)
}

// SetIndex sets an index on key. Building an index can take a long time.
func (db *levelDB) SetIndex(table, key string) error {
	return errors.New("SetIndex not implemented on LevelDB.")
}

// Terminate a database session (only run if nothing is using the database).
func (db *levelDB) Close() error {
	return db.dB.Close()
}

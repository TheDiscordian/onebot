package main

import (
	"errors"
	"fmt"
	. "github.com/TheDiscordian/onebot/loggers"
	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDB struct {
	path    string
	levelDB *leveldb.DB
}

func OpenLevelDB(path string) *LevelDB {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		Error.Panicln("Error opening levelDB database:", err)
	}
	return &LevelDB{path: path, levelDB: db}
}

// Get retrieves value by key directly
func (db *LevelDB) Get(table, key string) ([]byte, error) {
	data, err := db.levelDB.Get([]byte(fmt.Sprintf("%s.%s", table, key)), nil)
	return data, err
}

// Searches for key in field, containing key (IE: field:'username', key:'admin'), using an index if exists. Can be very
// slow without an index.
func (db *LevelDB) Search(table, field, key string) ([]byte, error) {
	if field == "_id" {
		return db.Get(table, key)
	} else {
		return nil, errors.New("Searching on LevelDB not yet implemented.")
	}
}

// Inserts value into key, erasing any potential previous value.
func (db *LevelDB) Put(table, key string, value []byte) error {
	return db.levelDB.Put([]byte(key), value, nil)
}

// SetIndex sets an index on key. Building an index can take a long time.
func (db *LevelDB) SetIndex(table, key string) error {
	return errors.New("SetIndex not implemented on LevelDB.")
}

// Terminate a database session (only run if nothing is using the database).
func (db *LevelDB) Close() error {
	return db.levelDB.Close()
}

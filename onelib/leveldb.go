// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package onelib

import (
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
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
func (db *levelDB) Get(table, key string) (map[string]interface{}, error) {
	return nil, errors.New("Get not implemented on LevelDB.")
}

// Retrieve a string stored with PutString.
func (db *levelDB) GetString(table, key string) (string, error) {
	data, err := db.dB.Get([]byte(fmt.Sprintf("%s.%s", table, key)), nil)
	return string(data), err
}

// Retrieve an integer stored with PutInt.
func (db *levelDB) GetInt(table, key string) (int, error) {
	data, err := db.dB.Get([]byte(fmt.Sprintf("%s.%s", table, key)), nil)
	if err != nil {
		return 0, err
	}
	var i int
	i, err = strconv.Atoi(string(data))
	return i, err
}

// Retrieve an object stored with PutObj.
func (db *levelDB) GetObj(table, key string, obj interface{}) error {
	data, err := db.dB.Get([]byte(fmt.Sprintf("%s.%s", table, key)), nil)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(data, obj)
	return err
}

// Searches for key in field, containing key (IE: field:'username', key:'admin'), using an index if exists. Can be very
// slow without an index.
func (db *levelDB) Search(table, field, key string) (map[string]interface{}, error) {
	if field == "_id" {
		return db.Get(table, key)
	}
	return nil, errors.New("Searching on LevelDB not yet implemented.")
}

// Inserts value into key, erasing any potential previous value.
func (db *levelDB) Put(table string, data map[string]interface{}) ([]byte, error) {
	return nil, errors.New("Put not implemented on LevelDB.")
}

// Inserts text at location "key" for retrieval via GetString
func (db *levelDB) PutString(table, key, text string) error {
	return db.dB.Put([]byte(fmt.Sprintf("%s.%s", table, key)), []byte(text), nil)
}

// Inserts an integer at location "key" for retrieval via GetInt
func (db *levelDB) PutInt(table, key string, i int) error {
	return db.dB.Put([]byte(fmt.Sprintf("%s.%s", table, key)), []byte(strconv.Itoa(i)), nil)
}

// Inserts an object at location "key" for retrieval via GetObj
func (db *levelDB) PutObj(table, key string, obj interface{}) error {
	data, err := bson.Marshal(obj)
	if err != nil {
		return err
	}
	return db.dB.Put([]byte(fmt.Sprintf("%s.%s", table, key)), data, nil)
}

// Removes an object at location "key"
func (db *levelDB) Remove(table, key string) error {
	return db.dB.Delete([]byte(fmt.Sprintf("%s.%s", table, key)), nil)
}

// SetIndex sets an index on field. Building an index can take a long time. On LevelDB Index *must* be unique, or will
// overwrite.
func (db *levelDB) SetIndex(table, field string) error {
	return errors.New("SetIndex not implemented on LevelDB.")
}

// Terminate a database session (only run if nothing is using the database).
func (db *levelDB) Close() error {
	return db.dB.Close()
}

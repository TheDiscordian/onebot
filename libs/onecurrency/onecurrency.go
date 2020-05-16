// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package onecurrency

import (
	"errors"
	"github.com/TheDiscordian/onebot/onelib"
	"sync"
)

// TODO support aliases. Have a simple interface for users to specify they have an alias. If a UUID is an alias, point
// to the master account, and use that currency. Aliases cannot be undone by user.

// DB_TABLE the table to use for onecurrency things
const DB_TABLE = "onecurrency_money"

var Currency map[string]*currencyStore

type currencyStore struct {
	storeMap map[string]*CurrencyObject
	lock     *sync.RWMutex
}

func InitCurrency(currency string) {
	cs := new(currencyStore)
	cs.storeMap = make(map[string]*CurrencyObject, 2)
	cs.lock = new(sync.RWMutex)
	Currency[currency] = cs
}

func (cs *currencyStore) save(uuid string, cObj *CurrencyObject) {
	onelib.Db.PutObj(DB_TABLE, uuid, cObj)
}

func (cs *currencyStore) load(uuid string) *CurrencyObject {
	var err error
	dbObj := new(CurrencyObject)
	err = onelib.Db.GetObj(DB_TABLE, uuid, dbObj)
	if err != nil {
		panic(err)
	}
	if dbObj == nil {
		return new(CurrencyObject)
	}
	return dbObj
}

// Get returns a pointer to a copy of the data in a target CurrencyObject.
func (cs *currencyStore) Get(uuid string) (cObj *CurrencyObject) {
	if cs == nil {
		return nil
	}
	cs.lock.RLock()
	if cs.storeMap[uuid] != nil {
		cObj = new(CurrencyObject)
		*cObj = *cs.storeMap[uuid]
	}
	cs.lock.RUnlock()
	return cObj
}

func (cs *currencyStore) set(uuid string, cObj *CurrencyObject) {
	if cs.storeMap == nil {
		cs.storeMap = make(map[string]*CurrencyObject, 2)
	}
	cs.storeMap[uuid] = cObj
}

// Set sets the target uuid to the desired CurrencyObject. Do not use the object again after calling this.
func (cs *currencyStore) Set(uuid string, cObj *CurrencyObject) {
	if cs == nil {
		return
	}
	cs.lock.Lock()
	cs.set(uuid, cObj)
	cs.save(uuid, cObj)
	cs.lock.Unlock()
}

// Add adds quantity to the Quantity of the stored CurrencyObject. It's atomic-safe, and should be quite fast, if used responsibly (IE: Make sure an object exists to add to, before calling this).
func (cs *currencyStore) Add(uuid string, displayName string, quantity int) (int, error) {
	if cs == nil {
		return 0, errors.New("selected currency type doesn't exist")
	}
	cs.lock.Lock()
	cObj := cs.storeMap[uuid]
	if cObj == nil {
		cObj = cs._new(uuid, displayName)
	}
	newQuantity := quantity + cObj.Quantity
	cObj.Quantity = newQuantity
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return newQuantity, nil
}

// Multiply multuplies quantity with the Quantity of the stored CurrencyObject. It's atomic-safe, and should be quite fast. It rounds down.
// TODO investigate if we'd rather just use float64 on the result, too.
func (cs *currencyStore) Multiply(uuid string, displayName string, quantity float64) (int, error) {
	if cs == nil {
		return 0, errors.New("selected currency type doesn't exist")
	}
	cs.lock.Lock()
	cObj := cs.storeMap[uuid]
	if cObj == nil { // race condition on first occurance of a user. Best way to mitigate this is to make sure cObj is never nil. This should be considered a fallback, code you want to avoid.
		cObj = cs._new(uuid, displayName)
	}
	newQuantity := int(quantity * float64(cObj.Quantity))
	cObj.Quantity = newQuantity
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return newQuantity, nil
}

// slightly different than New. Doesn't touch mutex locks, doesn't make a copy, doesn't save to database.
func (cs *currencyStore) _new(uuid string, displayName string) *CurrencyObject {
	cObj := cs.load(uuid)
	cObj.DisplayName = displayName
	cs.set(uuid, cObj)
	return cObj
}

// New creates, stores, and returns a copy of a CurrencyObject.
func (cs *currencyStore) New(uuid string, displayName string) *CurrencyObject {
	cObj := cs.load(uuid)
	cObj.DisplayName = displayName
	cObjCopy := *cObj
	cs.Set(uuid, cObj)
	return &cObjCopy
}

type CurrencyObject struct {
	DisplayName  string `bson:"dN"` // DisplayName of the user
	Quantity     int    `bson:"q"`  // Quantity of currency
	BankQuantity int    `bson:"bQ"` // Quantity of currency in bank
	Aliases      string `bson:"a"`  // User aliases
}

func init() {
	Currency = make(map[string]*currencyStore, 1)
}

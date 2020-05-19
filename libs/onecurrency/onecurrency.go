// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package onecurrency

import (
	"errors"
	"github.com/TheDiscordian/onebot/onelib"
	"sync"
)

// TODO alias needs to confirm with user. Timers need to respect alias.

// DB_TABLE the table to use for onecurrency things
const DB_TABLE = "onecurrency_money"

// TODO a "GetAll(currency string) []*CurrencyObject" feature. Returns sorted list of Quantity+BankQuantity descending. Useful for leaderboards, mass bonuses.
var Currency map[string]*currencyStore

type currencyStore struct {
	storeMap map[onelib.UUID]*CurrencyObject
	lock     *sync.Mutex
}

func InitCurrency(currency string) {
	cs := new(currencyStore)
	cs.storeMap = make(map[onelib.UUID]*CurrencyObject, 2)
	cs.lock = new(sync.Mutex)
	Currency[currency] = cs
}

// TODO put currency type into key so keys don't overwrite (Ex: CURRENCY_TYPE+DB_TABLE). Do this by storing unexported
// currency type in each CurrencyObject. Allow plugins to use UUID as they wish.
func (cs *currencyStore) save(uuid onelib.UUID, cObj *CurrencyObject) {
	if err := onelib.Db.PutObj(DB_TABLE, string(uuid), cObj); err != nil {
		onelib.Error.Println("PutObj Error:", err)
	}
}

func (cs *currencyStore) load(uuid onelib.UUID) *CurrencyObject {
	dbObj := new(CurrencyObject)
	err := onelib.Db.GetObj(DB_TABLE, string(uuid), dbObj)
	if err != nil {
		onelib.Error.Println("GetObj Error:", err)
		return nil
	}
	return dbObj
}

// Alias sets an alias to the target UUID.
func (cs *currencyStore) Alias(uuid, targetUUID onelib.UUID) error {
	if cs == nil {
		return nil
	}
	cs.lock.Lock()
	tcObj := cs.storeMap[targetUUID]
	if tcObj == nil {
		tcObj = cs.load(targetUUID)
	}
	if tcObj == nil {
		cs.lock.Unlock()
		return errors.New("target UUID doesn't resolve to a known user")
	}
	cObj := cs.storeMap[uuid]
	if cObj == nil {
		cObj = cs._new(uuid)
	}
	cObj.Alias = string(targetUUID)
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return nil
}

// UnAlias sets removes an alias from the UUID.
func (cs *currencyStore) UnAlias(uuid onelib.UUID) {
	if cs == nil {
		return
	}
	cs.lock.Lock()
	cObj := cs.storeMap[uuid]
	if cObj == nil {
		cObj = cs._new(uuid)
	}
	cObj.Alias = ""
	cs.save(uuid, cObj)
	cs.lock.Unlock()
}

func (cs *currencyStore) lookupUUID(uuid onelib.UUID) (cObj *CurrencyObject) {
	cObj = cs.storeMap[uuid]
	if cObj == nil {
		cObj = cs.load(uuid)
		cs.set(uuid, cObj)
	}
	return
}

// _get returns a pointer to the data in a target CurrencyObject. It does not respect mutex locks. Result will never be
// nil.
func (cs *currencyStore) _get(uuid onelib.UUID) (ruuid onelib.UUID, cObj *CurrencyObject) {
	cObj = cs.storeMap[uuid]
	ruuid = uuid
	if cObj == nil {
		cObj = cs._new(uuid)
	}
	if cObj.Alias != "" {
		tcObj := cs.lookupUUID(onelib.UUID(cObj.Alias))
		if tcObj == nil {
			onelib.Error.Println("Tried to lookup alias", cObj.Alias, "of", uuid, "but failed...")
		} else {
			ruuid = onelib.UUID(cObj.Alias)
			cObj = tcObj
		}
	}
	return
}

// Get returns a pointer to a copy of the data in a target CurrencyObject.
func (cs *currencyStore) Get(uuid onelib.UUID) (ruuid onelib.UUID, cObj *CurrencyObject) {
	if cs == nil {
		return
	}
	cs.lock.Lock()
	ruuid = uuid
	tcObj := cs.lookupUUID(uuid)
	if tcObj != nil {
		cObj = new(CurrencyObject)
		if tcObj.Alias != "" {
			acObj := cs.lookupUUID(onelib.UUID(tcObj.Alias))
			if acObj == nil {
				onelib.Error.Println("Tried to lookup alias", tcObj.Alias, "of", uuid, "but failed...")
				*cObj = *cs.storeMap[uuid]
			} else {
				ruuid = onelib.UUID(tcObj.Alias)
				*cObj = *acObj
			}
		} else {
			*cObj = *cs.storeMap[uuid]
		}
	}
	cs.lock.Unlock()
	return
}

func (cs *currencyStore) set(uuid onelib.UUID, cObj *CurrencyObject) {
	if cs.storeMap == nil {
		cs.storeMap = make(map[onelib.UUID]*CurrencyObject, 2)
	}
	cs.storeMap[uuid] = cObj
}

// Set sets the target uuid to the desired CurrencyObject. Will follow aliases. Do not use the object again after calling this.
func (cs *currencyStore) Set(uuid onelib.UUID, cObj *CurrencyObject) {
	if cs == nil {
		return
	}
	cs.lock.Lock()
	tcObj := cs.storeMap[uuid]
	if tcObj != nil && tcObj.Alias != "" {
		uuid = onelib.UUID(tcObj.Alias)
		cObj = cs.lookupUUID(uuid)
		if cObj == nil {
			onelib.Error.Println("Returned alias is nil:", tcObj.Alias)
		}
	}
	cs.set(uuid, cObj)
	cs.save(uuid, cObj)
	cs.lock.Unlock()
}

// Add adds quantity to the Quantity of the stored CurrencyObject. It's atomic-safe, and should be quite fast.
func (cs *currencyStore) Add(uuid onelib.UUID, quantity, bankQuantity int) (int, int, error) {
	if cs == nil {
		return 0, 0, errors.New("selected currency type doesn't exist")
	}
	cs.lock.Lock()
	var cObj *CurrencyObject
	uuid, cObj = cs._get(uuid)
	newQuantity := quantity + cObj.Quantity
	cObj.Quantity = newQuantity
	newBankQuantity := bankQuantity + cObj.BankQuantity
	cObj.BankQuantity = newBankQuantity
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return newQuantity, newBankQuantity, nil
}

// Multiply multuplies quantity with the Quantity of the stored CurrencyObject. It's atomic-safe, and should be quite fast. It rounds down.
// TODO investigate if we'd rather just use float64 on the result, too.
func (cs *currencyStore) Multiply(uuid onelib.UUID, quantity, bankQuantity float64) (int, int, error) {
	if cs == nil {
		return 0, 0, errors.New("selected currency type doesn't exist")
	}
	cs.lock.Lock()
	var cObj *CurrencyObject
	uuid, cObj = cs._get(uuid)
	newQuantity := int(quantity * float64(cObj.Quantity))
	cObj.Quantity = newQuantity
	newBankQuantity := int(bankQuantity * float64(cObj.BankQuantity))
	cObj.BankQuantity = newBankQuantity
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return newQuantity, newBankQuantity, nil
}

// _deposit attempts to do a deposit operation. Will only work if it results in Quantity being 0 or greater. Doesn't
// respect mutex locks.
func (cs *currencyStore) _deposit(cObj *CurrencyObject, uuid onelib.UUID, quantity int) error {
	if cObj.Quantity < quantity {
		return errors.New("insufficient funds")
	}
	cObj.Quantity -= quantity
	cObj.BankQuantity += quantity
	return nil
}

// Deposit attempts to do a deposit operation. Will only work if it results in Quantity being 0 or greater.
func (cs *currencyStore) Deposit(uuid onelib.UUID, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if cs == nil {
		return errors.New("selected currency type doesn't exist")
	}
	cs.lock.Lock()
	var cObj *CurrencyObject
	uuid, cObj = cs._get(uuid)
	err := cs._deposit(cObj, uuid, quantity)
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return err
}

// DepositAll attempts to deposit all of Quantity into BankQuantity.
func (cs *currencyStore) DepositAll(uuid onelib.UUID) (int, error) {
	if cs == nil {
		return 0, errors.New("selected currency type doesn't exist")
	}
	cs.lock.Lock()
	var cObj *CurrencyObject
	uuid, cObj = cs._get(uuid)
	all := cObj.Quantity
	if all <= 0 {
		cs.lock.Unlock()
		return 0, errors.New("insufficient funds")
	}
	err := cs._deposit(cObj, uuid, all)
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return all, err
}

// _withdraw attempts to do a withdraw operation. Will only work if it results in BankQuantity being 0 or greater.
// Doesn't respect mutex locks.
func (cs *currencyStore) _withdraw(cObj *CurrencyObject, uuid onelib.UUID, quantity int) error {
	if cObj.BankQuantity < quantity {
		return errors.New("insufficient funds")
	}
	cObj.BankQuantity -= quantity
	cObj.Quantity += quantity
	return nil
}

// Withdraw attempts to do a withdraw operation. Will only work if it results in BankQuantity being 0 or greater.
func (cs *currencyStore) Withdraw(uuid onelib.UUID, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if cs == nil {
		return errors.New("selected currency type doesn't exist")
	}
	cs.lock.Lock()
	var cObj *CurrencyObject
	uuid, cObj = cs._get(uuid)
	err := cs._withdraw(cObj, uuid, quantity)
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return err
}

// DepositAll attempts to deposit all of BankQuantity into ,balQuantity.
func (cs *currencyStore) WithdrawAll(uuid onelib.UUID) (int, error) {
	if cs == nil {
		return 0, errors.New("selected currency type doesn't exist")
	}
	cs.lock.Lock()
	var cObj *CurrencyObject
	uuid, cObj = cs._get(uuid)
	all := cObj.BankQuantity
	if all <= 0 {
		cs.lock.Unlock()
		return 0, errors.New("insufficient funds")
	}
	err := cs._withdraw(cObj, uuid, all)
	cs.save(uuid, cObj)
	cs.lock.Unlock()
	return all, err
}

// slightly different than New. Doesn't touch mutex locks, doesn't make a copy, doesn't save to database.
func (cs *currencyStore) _new(uuid onelib.UUID) *CurrencyObject {
	cObj := cs.load(uuid)
	if cObj == nil {
		cObj = new(CurrencyObject)
	}
	cs.set(uuid, cObj)
	return cObj
}

// New creates, stores, and returns a copy of a CurrencyObject.
func (cs *currencyStore) New(uuid onelib.UUID) *CurrencyObject {
	cObj := cs.load(uuid)
	if cObj == nil {
		cObj = new(CurrencyObject)
	}
	cObjCopy := *cObj
	cs.Set(uuid, cObj)
	return &cObjCopy
}

type CurrencyObject struct {
	Quantity     int    `bson:"q"`  // Quantity of currency
	BankQuantity int    `bson:"bQ"` // Quantity of currency in bank
	Alias        string `bson:"a"`  // User alias, UUID
	currencyType string
}

func init() {
	Currency = make(map[string]*currencyStore, 1)
}

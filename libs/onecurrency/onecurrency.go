// Copyright (c) 2020, The OneBot Contributors. All rights reserved.

package onecurrency

import (
	"errors"
	"fmt"
	"github.com/TheDiscordian/onebot/onelib"
	"sort"
	"sync"
	"time"
)

// DB_TABLE the table to use for onecurrency things
const DB_TABLE = "onecurrency_money"

// Currency is an object used for accessing currency objects managed by onecurrency.
// TODO a "GetAll(currency string) []*CurrencyObject" feature. Returns sorted list of Quantity+BankQuantity descending. Useful for leaderboards, mass bonuses.
var Currency *currencyStore

type currencyStore struct {
	locationMap map[onelib.UUID]*LocationObject
	userMap     map[onelib.UUID]*UserObject
	lock        *sync.RWMutex
	saveTimer   map[onelib.UUID]time.Time // key: [location UUID][currency type]
}

// UserCurrencyObject is a CurrencyObject that also has a `UUID` variable.
type UserCurrencyObject struct {
	UUID onelib.UUID
	*CurrencyObject
}

// GetAll copies every single user from "currency", and returns sorted list by Quantity+BankQuantity descending. Includes aliased users (which can't really be looked up).
func (cs *currencyStore) GetAll(currency string, location onelib.UUID) (out []*UserCurrencyObject) {
	cs.Get(currency, location, "") // FIXME this is clearly a jank thing here, this is used to ensure the location is loaded, we don't care about the result, it will load the location if it really exists. Ideally we'll move all this work into one routine.
	cs.lock.RLock()
	if cs.locationMap[location] == nil {
		cs.lock.RUnlock()
		return nil
	}
	out = make([]*UserCurrencyObject, 0, len(cs.locationMap[location].Currency[currency]))
	for key, value := range cs.locationMap[location].Currency[currency] {
		out = append(out, &UserCurrencyObject{UUID: key, CurrencyObject: &CurrencyObject{Quantity: value.Quantity, BankQuantity: value.BankQuantity, DisplayName: value.DisplayName}}) // copy
	}
	cs.lock.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Quantity+out[i].BankQuantity > out[j].Quantity+out[j].BankQuantity }) // sort, descending
	return
}

// saveLocation saves a LocationObject, it does NOT update a UserObject, if a user would be receiving new currencies. A
// save is not guaranteed to happen, but will most likely happen. Saves are intentionally delayed as the in-memory map
// is very fast, but the storage objects are very large. We ignore most save requests as requests usually happen in
// clusters anyways, so multiple saves isn't necessary.
func (cs *currencyStore) saveLocation(location onelib.UUID) {
	if time.Since(cs.saveTimer[location]) < time.Second*6 { // this value may need to be customized if under heavy load
		onelib.Debug.Println("Skipped save...") // FIXME perhaps if a save doesn't occur for a while, save again, just in case
		return
	}
	if err := onelib.Db.PutObj(DB_TABLE, "L"+string(location), cs.locationMap[location]); err != nil {
		onelib.Error.Println("PutObj Error:", err)
		return
	}
	cs.saveTimer[location] = time.Now()
}

// load retrieves a location object using location.
func (cs *currencyStore) loadLocation(location onelib.UUID) *LocationObject {
	dbObj := new(LocationObject)
	err := onelib.Db.GetObj(DB_TABLE, "L"+string(location), dbObj)
	if err != nil {
		onelib.Error.Println("GetObj Error:", err)
		return nil
	}
	return dbObj
}

func (cs *currencyStore) saveUser(uuid onelib.UUID) {
	if err := onelib.Db.PutObj(DB_TABLE, "U"+string(uuid), cs.userMap[uuid]); err != nil {
		onelib.Error.Println("PutObj Error:", err)
		return
	}
}

func (cs *currencyStore) loadUser(uuid onelib.UUID) *UserObject {
	dbObj := new(UserObject)
	err := onelib.Db.GetObj(DB_TABLE, "U"+string(uuid), dbObj)
	if err != nil {
		onelib.Error.Println("GetObj Error:", err)
		return nil
	}
	return dbObj
}

// Alias sets an alias to the target UUID.
func (cs *currencyStore) Alias(uuid, targetUUID onelib.UUID) error {
	cs.lock.Lock()
	tuObj := cs.userMap[targetUUID]
	if tuObj == nil {
		tuObj = cs.loadUser(targetUUID)
		cs.userMap[targetUUID] = tuObj
		if tuObj == nil {
			cs.lock.Unlock()
			return errors.New("target UUID doesn't resolve to a known user")
		}
	}
	if tuObj.Alias != onelib.UUID("") {
		cs.lock.Unlock()
		return errors.New("target UUID has an alias")
	}
	uObj := cs.userMap[uuid]
	if uObj == nil {
		uObj = cs.loadUser(uuid)
		cs.userMap[uuid] = uObj
		if uObj == nil {
			uObj = new(UserObject)
			cs.userMap[uuid] = uObj
			cs.userMap[uuid].Currencies = make(map[onelib.UUID][]string, 1)
		}
	}
	uObj.Alias = targetUUID
	cs.saveUser(uuid)
	cs.lock.Unlock()
	return nil
}

// UnAlias sets removes an alias from the UUID.
func (cs *currencyStore) UnAlias(uuid onelib.UUID) {
	cs.lock.Lock()
	uObj := cs.userMap[uuid]
	if uObj == nil {
		uObj = cs.loadUser(uuid)
		cs.userMap[uuid] = uObj
		if uObj == nil {
			cs.lock.Unlock()
			return
		}
	}
	uObj.Alias = ""
	cs.saveUser(uuid)
	cs.lock.Unlock()
}

func (cs *currencyStore) fetchAlias(uuid onelib.UUID) (err error) {
	tuObj := cs.userMap[uuid]
	if tuObj == nil {
		tuObj = cs.loadUser(uuid)
		if tuObj == nil {
			err = fmt.Errorf("alias lookup %s failed", uuid)
		} else {
			cs.userMap[uuid] = tuObj
		}
	}
	return
}

// assumes read lock.
func (cs *currencyStore) lookupAlias(uuid onelib.UUID) (err error) {
	tuObj := cs.userMap[uuid]
	if tuObj == nil {
		tuObj = cs.loadUser(uuid) // FIXME this db load isn't ever saved to a map
		if tuObj == nil {
			err = fmt.Errorf("alias lookup %s failed", uuid)
		}
	}
	return
}

// _get returns a pointer to the data in a target CurrencyObject. It assumes it's in a read lock. ruuid is always set,
// but cObj is only set if both the location uuid and user uuid point to valid locations. err being set doesn't mean
// cObj didn't get set (like on a failed alias lookup). If alias lookup failed, and the calling uuid doesn't have the
// currency, then an error will be returned along the lines of "user <UUID> doesn't have that currency".
// Will NOT load the location data from the DB, but can fetch aliased users from the DB (though it won't map them).
func (cs *currencyStore) _get(currency string, location, uuid onelib.UUID) (ruuid onelib.UUID, cObj *CurrencyObject, err error) {
	ruuid = uuid
	lObj := cs.locationMap[location]
	if lObj == nil {
		err = errors.New("location object is nil")
		return
	}
	uObj := cs.userMap[ruuid]
	if uObj == nil {
		err = errors.New("user object is nil")
		return
	}

	if uObj.Alias != "" {
		onelib.Debug.Println("Alias not empty")
		err = cs.lookupAlias(uObj.Alias)
		if err == nil {
			ruuid = uObj.Alias
		}
	}
	cObj = lObj.Currency[currency][ruuid]
	if cObj == nil {
		err = fmt.Errorf("user %s doesn't have that currency", ruuid)
	}
	return
}

// same as _get, but it expects the UserObject for the uuid as well
func (cs *currencyStore) __get(currency string, location, uuid onelib.UUID, uObj *UserObject) (ruuid onelib.UUID, cObj *CurrencyObject, err error) {
	ruuid = uuid
	lObj := cs.locationMap[location]
	if lObj == nil {
		err = errors.New("location object is nil")
		return
	}

	if uObj.Alias != "" {
		onelib.Debug.Println("Alias not empty")
		err = cs.lookupAlias(uObj.Alias)
		if err == nil {
			ruuid = uObj.Alias
		}
	}
	cObj = lObj.Currency[currency][ruuid]
	if cObj == nil {
		err = fmt.Errorf("user %s doesn't have that currency", ruuid)
	}
	return
}

// try to get a currency object, loading the location from the DB if necessary. This method assumes write access. This
// method will load UUID if not loaded. Look at this function as "check if exists anywhere, if so, return"
func (cs *currencyStore) fetchCurrencyObject(currency string, location, uuid onelib.UUID) (ruuid onelib.UUID, cObj *CurrencyObject, err error) {
	ruuid = uuid
	if cs.locationMap[location] == nil {
		cs.locationMap[location] = cs.loadLocation(location)
		if cs.locationMap[location] == nil {
			err = errors.New("location doesn't exist in DB")
			return
		}
	}
	if cs.userMap[uuid] == nil {
		cs.userMap[uuid] = cs.loadUser(uuid)
	}
	ruuid, cObj, err = cs._get(currency, location, uuid)
	return
}

// Get returns a pointer to a copy of the data in a target CurrencyObject.
// WILL load the location data from the DB.
func (cs *currencyStore) Get(currency string, location, uuid onelib.UUID) (ruuid onelib.UUID, cObj *CurrencyObject, err error) {
	var tcObj *CurrencyObject
	cObj = new(CurrencyObject)
	cs.lock.RLock()
	ruuid, tcObj, err = cs._get(currency, location, uuid)
	if err != nil {
		if errTxt := err.Error(); errTxt == "location object is nil" { // location hasn't been loaded, try to remedy that...
			cs.lock.RUnlock() // race ... (resolved by another check for location object presence in fetchCurrencyObject)
			cs.lock.Lock()
			ruuid, tcObj, err = cs.fetchCurrencyObject(currency, location, uuid)
			if tcObj != nil { // hopefully after all this, we have data ...
				*cObj = *tcObj
			}
			cs.lock.Unlock()
			return
		} else if errTxt == "user object is nil" { // okay, we don't have a user object, those are small so we'll just load a temporary object, if possible.
			uObj := cs.loadUser(uuid)
			if uObj != nil {
				ruuid, tcObj, err = cs.__get(currency, location, uuid, uObj)
			}
		}
	}
	if tcObj != nil {
		*cObj = *tcObj
	}
	cs.lock.RUnlock()
	return
}

func indexStrings(s []string, target string) (out int) {
	out = -1
	for index, ss := range s {
		if ss == target {
			out = index
			return
		}
	}
	return
}

// doesn't resolve aliases, saves user, but not location
func (cs *currencyStore) set(currency string, location, uuid onelib.UUID, cObj *CurrencyObject) {
	lObj := cs.locationMap[location]
	if lObj == nil { // no location in map, try to load from db...
		lObj = cs.loadLocation(location)
		cs.locationMap[location] = lObj
	}
	if lObj == nil { // location doesn't exists at all, make a new one...
		lObj = new(LocationObject)
		lObj.Currency = make(map[string]map[onelib.UUID]*CurrencyObject, 1)
		lObj.Currency[currency] = make(map[onelib.UUID]*CurrencyObject, 2)
		cs.locationMap[location] = lObj
	}
	lObj.Currency[currency][uuid] = cObj
	cs.saveLocation(location)

	uObj := cs.userMap[uuid]
	if uObj == nil { // no user in map, try to load from db...
		uObj = cs.loadUser(uuid)
		cs.userMap[uuid] = uObj
	}
	if uObj == nil { // user doesn't exist at all, make a new one...
		uObj = new(UserObject)
		uObj.Currencies = make(map[onelib.UUID][]string, 1)
		cs.userMap[uuid] = uObj
	}

	if len(uObj.Currencies[location]) == 0 { // does the user have any currencies?
		uObj.Currencies[location] = []string{currency}
		cs.userMap[uuid] = uObj
		cs.saveUser(uuid)
	} else if indexStrings(uObj.Currencies[location], currency) == -1 { // the user has currency, do they have ours?
		uObj.Currencies[location] = append(uObj.Currencies[location], currency)
		cs.userMap[uuid] = uObj
		cs.saveUser(uuid)
	}
}

// Set sets the target uuid to the desired CurrencyObject. Will follow aliases. Do not use the object again after calling this.
func (cs *currencyStore) Set(currency string, location, uuid onelib.UUID, cObj *CurrencyObject) {
	cs.lock.Lock()
	nuuid, _, err := cs._get(currency, location, uuid)
	if err != nil {
		onelib.Error.Println(err)
	}
	cs.set(currency, location, nuuid, cObj)
	cs.lock.Unlock()
}

// Add adds quantity to the Quantity of the stored CurrencyObject. It's atomic-safe, and should be quite fast.
func (cs *currencyStore) Add(currency string, location, uuid onelib.UUID, quantity, bankQuantity int) (newQuantity int, newBankQuantity int) {
	cs.lock.Lock()
	ruuid, cObj, _ := cs.fetchCurrencyObject(currency, location, uuid)
	newQuantity = quantity
	newBankQuantity = bankQuantity
	if cObj == nil {
		cObj = new(CurrencyObject)
		cObj.Quantity = quantity
		cObj.BankQuantity = bankQuantity
		cs.set(currency, location, ruuid, cObj)
	} else {
		newQuantity += cObj.Quantity
		cObj.Quantity = newQuantity
		newBankQuantity += cObj.BankQuantity
		cObj.BankQuantity = newBankQuantity
		cs.saveLocation(location)
	}
	cs.lock.Unlock()
	return newQuantity, newBankQuantity
}

// Multiply multuplies quantity with the Quantity of the stored CurrencyObject. It's atomic-safe, and should be quite fast. It rounds down.
// TODO investigate if we'd rather just use float64 on the result, too.
func (cs *currencyStore) Multiply(currency string, location, uuid onelib.UUID, quantity, bankQuantity float64) (newQuantity int, newBankQuantity int) {
	cs.lock.Lock()
	ruuid, cObj, _ := cs.fetchCurrencyObject(currency, location, uuid)
	if cObj == nil {
		cObj = new(CurrencyObject)
		cs.set(currency, location, ruuid, cObj)
	} else {
		cObj.Quantity = int(quantity * float64(cObj.Quantity))
		newQuantity = cObj.Quantity
		cObj.BankQuantity = int(bankQuantity * float64(cObj.BankQuantity))
		newBankQuantity = cObj.BankQuantity
		cs.saveLocation(location)
	}
	cs.lock.Unlock()
	return newQuantity, newBankQuantity
}

// _deposit attempts to do a deposit operation. Will only work if it results in Quantity being 0 or greater. Doesn't
// respect mutex locks.
func (cs *currencyStore) _deposit(cObj *CurrencyObject, quantity int) error {
	if cObj.Quantity < quantity {
		return errors.New("insufficient funds")
	}
	cObj.Quantity -= quantity
	cObj.BankQuantity += quantity
	return nil
}

// Deposit attempts to do a deposit operation. Will only work if it results in Quantity being 0 or greater.
func (cs *currencyStore) Deposit(currency string, location, uuid onelib.UUID, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	cs.lock.Lock()
	_, cObj, err := cs.fetchCurrencyObject(currency, location, uuid)
	if cObj == nil {
		cs.lock.Unlock()
		return err
	}
	err = cs._deposit(cObj, quantity)
	cs.saveLocation(location)
	cs.lock.Unlock()
	return err
}

// DepositAll attempts to deposit all of Quantity into BankQuantity.
func (cs *currencyStore) DepositAll(currency string, location, uuid onelib.UUID) (int, error) {
	cs.lock.Lock()
	_, cObj, err := cs.fetchCurrencyObject(currency, location, uuid)
	if cObj == nil {
		cs.lock.Unlock()
		return 0, err
	}
	all := cObj.Quantity
	if all <= 0 {
		cs.lock.Unlock()
		return 0, errors.New("insufficient funds")
	}
	err = cs._deposit(cObj, all)
	cs.saveLocation(location)
	cs.lock.Unlock()
	return all, err
}

// _withdraw attempts to do a withdraw operation. Will only work if it results in BankQuantity being 0 or greater.
// Doesn't respect mutex locks.
func (cs *currencyStore) _withdraw(cObj *CurrencyObject, quantity int) error {
	if cObj.BankQuantity < quantity {
		return errors.New("insufficient funds")
	}
	cObj.BankQuantity -= quantity
	cObj.Quantity += quantity
	return nil
}

// Withdraw attempts to do a withdraw operation. Will only work if it results in BankQuantity being 0 or greater.
func (cs *currencyStore) Withdraw(currency string, location, uuid onelib.UUID, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	cs.lock.Lock()
	_, cObj, err := cs.fetchCurrencyObject(currency, location, uuid)
	if cObj == nil {
		cs.lock.Unlock()
		return err
	}
	err = cs._withdraw(cObj, quantity)
	cs.saveLocation(location)
	cs.lock.Unlock()
	return err
}

// DepositAll attempts to deposit all of BankQuantity into ,balQuantity.
func (cs *currencyStore) WithdrawAll(currency string, location, uuid onelib.UUID) (int, error) {
	cs.lock.Lock()
	_, cObj, err := cs.fetchCurrencyObject(currency, location, uuid)
	if cObj == nil {
		cs.lock.Unlock()
		return 0, err
	}
	all := cObj.Quantity
	if all <= 0 {
		cs.lock.Unlock()
		return 0, errors.New("insufficient funds")
	}
	err = cs._withdraw(cObj, all)
	cs.saveLocation(location)
	cs.lock.Unlock()
	return all, err
}

func (cs *currencyStore) UpdateDisplayName(currency string, location, uuid onelib.UUID, name string) {
	cs.Get(currency, location, "") // FIXME this is clearly a jank thing here, this is used to ensure the location is loaded, we don't care about the result, it will load the location if it really exists. Ideally we'll move all this work into one routine.
	cs.lock.RLock()
	if cs.locationMap[location] == nil { // location doesn't exist
		cs.lock.RUnlock()
		return
	}
	if cs.locationMap[location].Currency[currency] == nil { // currency type doesn't exist
		cs.lock.RUnlock()
		return
	}
	if cs.locationMap[location].Currency[currency][uuid] == nil { // user isn't registered with that currency type
		cs.lock.RUnlock()
		return
	}
	if cs.locationMap[location].Currency[currency][uuid].DisplayName == name { // display name hasn't changed
		cs.lock.RUnlock()
		return
	}
	cs.lock.RUnlock()
	cs.lock.Lock()
	if cs.locationMap[location].Currency[currency][uuid].DisplayName != name { // make sure we didn't lose a race
		cs.locationMap[location].Currency[currency][uuid].DisplayName = name
		cs.saveLocation(location)
	}
	cs.lock.Unlock()
}

// LocationObject is an object representing a community, which stores all the currency values for all its users.
type LocationObject struct {
	Currency map[string]map[onelib.UUID]*CurrencyObject `bson:"c"` // key: [currencyType][user UUID]
}

// CurrencyObject is an object containing an amount of currency.
type CurrencyObject struct {
	Quantity     int    `bson:"q"`  // Quantity of currency
	BankQuantity int    `bson:"bQ"` // Quantity of currency in bank
	DisplayName  string `bson:"dN"` // DisplayName of the UUID who owns this CurrencyObject
}

// UserObject represents an account, and can be aliased to another UserObject via UUID. It stores all its known
// currencies for reverse-lookups.
type UserObject struct {
	Alias      onelib.UUID              `bson:"a"` // User alias, UUID
	Currencies map[onelib.UUID][]string `bson:"c"` // map of location UUIDs to a list of currency types
}

func init() {
	Currency = new(currencyStore)
	Currency.userMap = make(map[onelib.UUID]*UserObject, 2)
	Currency.locationMap = make(map[onelib.UUID]*LocationObject, 1)
	Currency.lock = new(sync.RWMutex)
	Currency.saveTimer = make(map[onelib.UUID]time.Time, 1)
}

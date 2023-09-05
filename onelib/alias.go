// Copyright (c) 2020-2022, The OneBot Contributors. All rights reserved.

package onelib

import (
	"errors"
	"fmt"
	"sync"
)

var Alias *userStore

type userStore struct {
	userMap map[UUID]*UserObject
	lock    *sync.RWMutex
}

func init() {
	Alias = &userStore{
		userMap: make(map[UUID]*UserObject, 1),
		lock:    new(sync.RWMutex),
	}
}

const (
	AliasTable = "onelib_aliases"
)

// UserObject represents an account, and can be aliased to another UserObject via UUID.
type UserObject struct {
	Alias UUID `bson:"a"` // User alias, UUID
}

func (us *userStore) saveUser(uuid UUID) {
	if err := Db.PutObj(AliasTable, "U"+string(uuid), us.userMap[uuid]); err != nil {
		Error.Println("PutObj Error:", err)
		return
	}
}

func (us *userStore) loadUser(uuid UUID) *UserObject {
	dbObj := new(UserObject)
	err := Db.GetObj(AliasTable, "U"+string(uuid), dbObj)
	if err != nil {
		if err.Error() != "leveldb: not found" {
			Error.Println("GetObj Error:", err)
		}
		return nil
	}
	return dbObj
}

// SetAlias sets an alias to the target UUID.
func (us *userStore) Set(uuid, targetUUID UUID) error {
	us.lock.Lock()
	tuObj := us.userMap[targetUUID]
	if tuObj == nil {
		tuObj = us.loadUser(targetUUID)
		us.userMap[targetUUID] = tuObj
		if tuObj == nil {
			tuObj = new(UserObject)
			us.userMap[targetUUID] = tuObj
		}
	}
	if tuObj.Alias != UUID("") {
		us.lock.Unlock()
		return errors.New("target UUID has an alias")
	}
	uObj := us.userMap[uuid]
	if uObj == nil {
		uObj = us.loadUser(uuid)
		us.userMap[uuid] = uObj
		if uObj == nil {
			uObj = new(UserObject)
			us.userMap[uuid] = uObj
		}
	}
	uObj.Alias = targetUUID
	us.saveUser(uuid)
	us.lock.Unlock()
	return nil
}

// UnAlias sets removes an alias from the UUID.
func (us *userStore) UnAlias(uuid UUID) {
	us.lock.Lock()
	uObj := us.userMap[uuid]
	if uObj == nil {
		uObj = us.loadUser(uuid)
		us.userMap[uuid] = uObj
		if uObj == nil {
			us.lock.Unlock()
			return
		}
	}
	uObj.Alias = ""
	us.saveUser(uuid)
	us.lock.Unlock()
}

// Get retrieves the alias pointing to the user (if any), erroring if the user has never had an alias.
func (us *userStore) Get(uuid UUID) (alias UUID, err error) {
	us.lock.RLock()
	tuObj := us.userMap[uuid]
	if tuObj == nil {
		tuObj = us.loadUser(uuid)
		if tuObj == nil {
			err = fmt.Errorf("alias lookup %s failed", uuid)
		} else {
			us.lock.RUnlock()
			us.lock.Lock()
			alias = tuObj.Alias
			if us.userMap[uuid] == nil {
				us.userMap[uuid] = tuObj
			}
			us.lock.Unlock()
			return
		}
	} else {
		alias = tuObj.Alias
	}
	us.lock.RUnlock()
	return
}

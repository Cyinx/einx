package mongodb

import (
	"errors"
	"fmt"
	"github.com/Cyinx/einx/component"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/slog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type M = bson.M
type ModuleEventer = module.ModuleEventer

const (
	Strong    = 1
	Monotonic = 2
)

type MongoDBMgr struct {
	session      *mgo.Session
	timeout      time.Duration
	dbcfg        *MongoDBInfo
	component_id ComponentID
	module       ModuleEventer
}

func NewMongoDBMgr(m ModuleEventer, dbcfg *MongoDBInfo, timeout time.Duration) *MongoDBMgr {
	return &MongoDBMgr{
		session:      nil,
		timeout:      timeout,
		dbcfg:        dbcfg,
		component_id: component.GenComponentID(),
		module:       m,
	}
}

func (this *MongoDBMgr) GetID() ComponentID {
	return this.component_id
}

func (this *MongoDBMgr) GetType() ComponentType {
	return component.COMPONENT_TYPE_DB_MONGODB
}

func (this *MongoDBMgr) Start() {
	var err error
	this.session, err = mgo.DialWithTimeout(this.dbcfg.String(), this.timeout)
	if err != nil {
		panic(err.Error())
		return
	}

	this.session.SetMode(mgo.Monotonic, true)
	slog.LogInfo("mongodb", "MongoDB Connect %v mongodb...success", this.dbcfg.String())
}

func (this *MongoDBMgr) Close() {
	if this.session != nil {
		this.session.DB("").Logout()
		this.session.Close()
		this.session = nil
		slog.LogInfo("mongodb", "Disconnect mongodb url: ", this.dbcfg.String())
	}
}

func (this *MongoDBMgr) RefreshSession() {
	this.session.Refresh()

}

func (db *MongoDBMgr) GetDbSession() *mgo.Session {
	return db.session
}

func (this *MongoDBMgr) SetMode(mode int, refresh bool) {
	status := mgo.Monotonic
	if mode == Strong {
		status = mgo.Strong
	} else {
		status = mgo.Monotonic
	}

	this.session.SetMode(status, refresh)
}

func (this *MongoDBMgr) Insert(collection string, doc interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)

	return c.Insert(doc)
}

func (this *MongoDBMgr) StrongInsert(collection string, doc interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	db_session.SetMode(mgo.Strong, true)

	c := db_session.DB("").C(collection)

	return c.Insert(doc)
}

func (this *MongoDBMgr) Update(collection string, cond interface{}, change interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)

	return c.Update(cond, bson.M{"$set": change})
}

func (this *MongoDBMgr) StrongUpdate(collection string, cond interface{}, change interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	db_session.SetMode(mgo.Strong, true)
	c := db_session.DB("").C(collection)

	return c.Update(cond, bson.M{"$set": change})
}

func (this *MongoDBMgr) UpdateInsert(collection string, cond interface{}, doc interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	_, err := c.Upsert(cond, bson.M{"$set": doc})
	if err != nil {
		slog.LogInfo("mongodb", "UpdateInsert failed collection is:%s. cond is:%v", collection, cond)
	}

	return err
}

func (this *MongoDBMgr) RemoveOne(collection string, cond_name string, cond_value int64) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	err := c.Remove(bson.M{cond_name: cond_value})
	if err != nil && err != mgo.ErrNotFound {
		slog.LogInfo("mongodb", "remove failed from collection:%s. name:%s-value:%d", collection, cond_name, cond_value)
	}

	return err

}

func (this *MongoDBMgr) RemoveOneByCond(collection string, cond interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	err := c.Remove(cond)

	if err != nil && err != mgo.ErrNotFound {
		slog.LogInfo("mongodb", "remove failed from collection:%s. cond :%v, err: %v.", collection, cond, err)
	}

	return err

}

func (this *MongoDBMgr) RemoveAll(collection string, cond interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	change, err := c.RemoveAll(cond)
	if err != nil && err != mgo.ErrNotFound {
		slog.LogInfo("mongodb", "MongoDBMgr RemoveAll failed : %s, %v", collection, cond)
		return err
	}
	slog.LogInfo("mongodb", "MongoDBMgr RemoveAll: %v, %v", change.Updated, change.Removed)
	return nil
}

func (this *MongoDBMgr) DBQueryOne(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	q := c.Find(cond)

	m := make(bson.M)
	if err := q.One(m); err != nil {
		if mgo.ErrNotFound != err {
			slog.LogInfo("mongodb", "DBFindOne query falied,return error: %v; name: %v.", err, collection)
		}
		return err
	}

	if nil != resHandler {
		return resHandler(m)
	}

	return nil

}

func (this *MongoDBMgr) StrongDBQueryOne(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	db_session.SetMode(mgo.Strong, true)

	c := db_session.DB("").C(collection)
	q := c.Find(cond)

	m := make(bson.M)
	if err := q.One(m); err != nil {
		if mgo.ErrNotFound != err {
			slog.LogInfo("mongodb", "DBFindOne query falied, return error: %v; name: %v.", err, collection)
		}
		return err
	}

	if nil != resHandler {
		return resHandler(m)
	}

	return nil

}

func (this *MongoDBMgr) DBQueryAll(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	q := c.Find(cond)

	if nil == q {
		return MONGODB_DBFINDALL_ERR
	}

	iter := q.Iter()
	m := make(bson.M)
	for iter.Next(m) == true {
		if nil != resHandler {
			err := resHandler(m)
			if err != nil {
				slog.LogInfo("mongodb", "resHandler error :%v!!!", err)
				return err
			}
		}
	}

	return nil

}

func (this *MongoDBMgr) StrongDBQueryAll(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	db_session.SetMode(mgo.Strong, true)

	c := db_session.DB("").C(collection)
	q := c.Find(cond)

	slog.LogInfo("mongodb", "[MongoDBMgr.DBFindAll] name:%s,query:%v", collection, cond)

	if nil == q {
		return MONGODB_DBFINDALL_ERR
	}
	iter := q.Iter()
	m := make(bson.M)
	for iter.Next(m) == true {
		if resHandler != nil {
			err := resHandler(m)
			if err != nil {
				slog.LogInfo("mongodb", "resHandler error :%v!!!", err)
				return err
			}
		}
	}

	return nil
}

func (this *MongoDBMgr) DeleteOne(collection string, cond interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()
	db_session.SetMode(mgo.Strong, true)
	c := db_session.DB("").C(collection)
	return c.Remove(cond)
}

func (this *MongoDBMgr) DeleteAll(collection string, cond interface{}) (int, error) {
	if this.session == nil {
		return 0, MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()
	db_session.SetMode(mgo.Strong, true)
	c := db_session.DB("").C(collection)
	changeInfo, err := c.RemoveAll(cond)
	if err != nil {
		return 0, err
	}
	return changeInfo.Removed, nil
}

package mongodb

import (
	"errors"
	"fmt"
	"github.com/Cyinx/einx/slog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type M = bson.M

const (
	Strong    = 1
	Monotonic = 2
)

var (
	MONGODB_SESSION_NIL_ERR = errors.New("MongoDBManager session nil.")
	MONGODB_NOTFOUND_ERR    = errors.New("not found!")
	MONGODB_DBFINDALL_ERR   = errors.New("MongoDBManager found error")
)

type MongoDBInfo struct {
	DbHost string
	DbPort int
	DbName string
	DbUser string
	DbPass string
}

func NewMongoDBInfo(host string, port int, name, user, pass string) *MongoDBInfo {
	return &MongoDBInfo{
		DbHost: host,
		DbPort: port,
		DbName: name,
		DbUser: user,
		DbPass: pass,
	}
}

func (this *MongoDBInfo) String() string {
	url := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		this.DbUser, this.DbPass, this.DbHost, this.DbPort, this.DbName)
	if this.DbUser == "" || this.DbPass == "" {
		url = fmt.Sprintf("mongodb://%s:%d/%s", this.DbHost, this.DbPort, this.DbName)
	}
	return url
}

type MongoDBManager struct {
	session *mgo.Session
	timeout time.Duration
	dbcfg   *MongoDBInfo
}

func NewMongoDBManager(dbcfg *MongoDBInfo, timeout time.Duration) *MongoDBManager {
	return &MongoDBManager{nil, timeout, dbcfg}
}

func (db *MongoDBManager) GetDbSession() *mgo.Session {
	return db.session
}

func (this *MongoDBManager) SetMode(mode int, refresh bool) {
	status := mgo.Monotonic
	if mode == Strong {
		status = mgo.Strong
	} else {
		status = mgo.Monotonic
	}

	this.session.SetMode(status, refresh)
}

func (this *MongoDBManager) OpenDB(set_index_func func(ms *mgo.Session)) error {

	var err error
	this.session, err = mgo.DialWithTimeout(this.dbcfg.String(), this.timeout)
	if err != nil {
		panic(err.Error())
	}

	this.session.SetMode(mgo.Monotonic, true)
	//set index
	if set_index_func != nil {
		set_index_func(this.session)
	}
	slog.LogInfo("mongodb", "MongoDB Connect %v mongodb...success", this.dbcfg.String())
	return nil
}

func (this *MongoDBManager) CloseDB() {
	if this.session != nil {
		this.session.DB("").Logout()
		this.session.Close()
		this.session = nil
		slog.LogInfo("mongodb", "Disconnect mongodb url: ", this.dbcfg.String())
	}
}

func (this *MongoDBManager) RefreshSession() {
	this.session.Refresh()

}

func (this *MongoDBManager) Insert(collection string, doc interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)

	return c.Insert(doc)
}

func (this *MongoDBManager) StrongInsert(collection string, doc interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	db_session.SetMode(mgo.Strong, true)

	c := db_session.DB("").C(collection)

	return c.Insert(doc)
}

func (this *MongoDBManager) Update(collection string, cond interface{}, change interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)

	return c.Update(cond, bson.M{"$set": change})
}

func (this *MongoDBManager) StrongUpdate(collection string, cond interface{}, change interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	db_session.SetMode(mgo.Strong, true)
	c := db_session.DB("").C(collection)

	return c.Update(cond, bson.M{"$set": change})
}

func (this *MongoDBManager) UpdateInsert(collection string, cond interface{}, doc interface{}) error {
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

func (this *MongoDBManager) RemoveOne(collection string, cond_name string, cond_value int64) error {
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

func (this *MongoDBManager) RemoveOneByCond(collection string, cond interface{}) error {
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

func (this *MongoDBManager) RemoveAll(collection string, cond interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	c := db_session.DB("").C(collection)
	change, err := c.RemoveAll(cond)
	if err != nil && err != mgo.ErrNotFound {
		slog.LogInfo("mongodb", "MongoDBManager RemoveAll failed : %s, %v", collection, cond)
		return err
	}
	slog.LogInfo("mongodb", "MongoDBManager RemoveAll: %v, %v", change.Updated, change.Removed)
	return nil
}

func (this *MongoDBManager) DBQueryOne(collection string, cond interface{}, resHandler func(bson.M) error) error {
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

func (this *MongoDBManager) StrongDBQueryOne(collection string, cond interface{}, resHandler func(bson.M) error) error {
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

func (this *MongoDBManager) DBQueryAll(collection string, cond interface{}, resHandler func(bson.M) error) error {
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

func (this *MongoDBManager) StrongDBQueryAll(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()

	db_session.SetMode(mgo.Strong, true)

	c := db_session.DB("").C(collection)
	q := c.Find(cond)

	slog.LogInfo("mongodb", "[MongoDBManager.DBFindAll] name:%s,query:%v", collection, cond)

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

func (this *MongoDBManager) DeleteOne(collection string, cond interface{}) error {
	if this.session == nil {
		return MONGODB_SESSION_NIL_ERR
	}

	db_session := this.session.Copy()
	defer db_session.Close()
	db_session.SetMode(mgo.Strong, true)
	c := db_session.DB("").C(collection)
	return c.Remove(cond)
}

func (this *MongoDBManager) DeleteAll(collection string, cond interface{}) (int, error) {
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

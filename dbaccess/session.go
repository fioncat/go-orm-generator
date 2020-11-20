package dbaccess

import (
	"fmt"
	"sync"

	"github.com/fioncat/go-gendb/dbaccess/dbtypes"
	"github.com/fioncat/go-gendb/dbaccess/mysql"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/store"
)

var conn *store.Conn

func SetConn(key string) error {
	var err error
	conn, err = store.GetConn(key)
	if err != nil {
		return err
	}
	if conn == nil {
		return errors.Fmt("can not find conn %s", key)
	}
	return nil
}

type Session interface {
	Connect(conn *store.Conn) error
	Desc(table string) (*dbtypes.Table, error)
}

var (
	sess Session
	mu   sync.Mutex
)

func getSession(dbType string) (Session, error) {
	mu.Lock()
	defer mu.Unlock()
	if conn == nil {
		return nil, errors.New("connection is not set")
	}
	if sess == nil {
		switch dbType {
		case "mysql":
			sess = &mysql.Session{}
		default:
			return nil, errors.Fmt("unknown db type %s", dbType)
		}
		err := sess.Connect(conn)
		if err != nil {
			return nil, errors.Trace("connect database", err)
		}
		log.Infof("connect to database '%s' success", conn.Key)
	}
	return sess, nil
}

func Desc(dbType, table string) (*dbtypes.Table, error) {
	if conn == nil {
		return nil, errors.New("connection is not set")
	}
	// try to get table struct from cache
	key := fmt.Sprintf("%s_%s", conn.Database, table)
	var cacheTable dbtypes.Table
	ok := store.GetCache(conn.Key, key, &cacheTable)
	if ok {
		log.Infof("Get table %s from local-cache", table)
		return &cacheTable, nil
	}

	// fetch table in database
	dbSess, err := getSession(dbType)
	if err != nil {
		return nil, err
	}
	descTable, err := dbSess.Desc(table)
	if err != nil {
		return nil, err
	}
	log.Infof("Get table %s from remote database", table)

	// save table to the local cache
	err = store.SaveCache(conn.Key, key, descTable)
	if err != nil {
		log.Errorf("cache: save %s for conn %s failed: %v",
			key, conn.Key, err)
	}
	return descTable, nil
}

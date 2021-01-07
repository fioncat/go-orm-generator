package rdb

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/fioncat/go-gendb/database/conn"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
)

// Session stores the connection to the remote database
// (RDB only) and some related operations of the database.
// Using this structure method can directly manipulate the
// database without having to care about the database
// connection.
// Different types of databases (such as mysql, SQLServer, etc.)
// can be received by Session. When in use, only need to
// care about the parameters of different operations, not the
// specific database type. But need to give the type during
// initialization.
type Session struct {
	// database connection configuration.
	cfg *conn.Config

	// database operations.
	oper dbOper

	// database connection object
	db *sql.DB

	// function to connect database
	connect ConnectFunc
}

// Desc is used to describe a data table. It will return
// some basic information of the data table, including
// table name, comments, table field information, etc.
// These data are returned in the form of the Table interface,
// see the documentation of the Table interface for details.
//
// Note that when the ENABLE_TABLE_CACHE variable is true,
// it means that the table cache will be turned on, and the
// table data will be read from the local disk first and converted
// to the Table interface to return. If the disk has no cache
// data, the table structure obtained from the database will
// be saved on the disk. The next time you call Desc on the
// same table, you can use the cache. The cache expiration
// time defaults to the TABLE_CACHE_TTL variable.
func (s *Session) Desc(tableName string) (Table, error) {
	if !ENABLE_TABLE_CACHE {
		return s.oper.Desc(s.db, tableName)
	}

	// try to fetch data from disk.
	key := fmt.Sprintf("cache_table_%s_%s_%s", s.cfg.Key,
		s.cfg.Database, tableName)
	table := getCacheTable(key)
	if table != nil {
		return table, nil
	}

	// cache no hit, get table from database
	// and save to cache
	var err error
	table, err = s.oper.Desc(s.db, tableName)
	if err != nil {
		return nil, err
	}
	saveCacheTable(key, table)

	return table, nil
}

// Check is used to check sql statement, "sql" means the
// statement to be checked, and "prepares" means the "?"
// parameter list of the statement. Check can check out the
// syntax error and warning information of the SQL statement
// (such as full table scan), which is returned in the form
// of the CheckResult interface. See the interface documentation
// for details.
func (s *Session) Check(sql string, prepares []interface{}) (CheckResult, error) {
	return s.oper.Check(s.db, sql, prepares)
}

// GoType is an auxiliary function, it does not connect to
// the database, but converts the database field type to the
// Go type.
// The conversion logic of different types of databases is
// different.
func (s *Session) GoType(sqlType string) string {
	return s.oper.ConvertType(sqlType)
}

// dbOper represents internal database operations, and
// different types of databases have different implementations.
type dbOper interface {
	// Init will be called when the Session is initialized,
	// passing the Session itself to dbOper, so that dbOper
	// can obtain a database connection during initialization
	// for some potential initialization operations.
	Init(sess *Session)

	// Desc describes the concrete realization of the data
	// table. This is different from Session.Desc, which does
	// not involve caching and directly manipulates the database.
	Desc(db *sql.DB, tableName string) (Table, error)

	// Check is the specific implementation of checking sql statement
	Check(db *sql.DB, sql string, prepares []interface{}) (CheckResult, error)

	// ConvertType is a specific implementation of converting
	// database type to Go type.
	ConvertType(sqlType string) string
}

// Table describes the information of the RDB data table.
// The specific implementation of this interface may be
// different for different types of databases, so it is
// given in the form of an interface.
type Table interface {
	// GetName returns the name of the data table.
	GetName() string

	// GetComment returns the comment of the data table.
	GetComment() string

	// Field extracts the corresponding field based on
	// the field name. If the field is not found, nil
	// is returned.
	Field(name string) Field

	// FieldNames returns all field names. Generally
	// used to traverse all fields of the table.
	FieldNames() []string
}

// Field represents the data table field, which is
// generally retrieved through the Table.Field.
type Field interface {
	// GetName returns the name of the data table field
	GetName() string

	// GetName returns the comment of the data table field
	GetComment() string

	// GetName returns the db type of the data table field
	GetType() string
}

// CheckResult represents the result of checking the sql
// statement.
// The specific implementation of this interface may be
// different for different types of databases, so it is
// given in the form of an interface.
type CheckResult interface {
	// GetErr returns the error of checking sql. If there
	// is no error, return nil.
	GetErr() error

	// GetWarns returns the warning message of checking sql.
	GetWarns() []string
}

// ConnectFunc represents a function to connect to the database.
// Different databases have different implementations.
type ConnectFunc func(cfg *conn.Config) (*sql.DB, error)

// Save the mapping of all currently supported database
// types and their Session (not yet connected).
var initSessM = make(map[string]*Session)

// Create a session that is not connected to the database,
// the session here cannot call the database operation method,
// otherwise it will panic
func newSess(oper dbOper, connFunc ConnectFunc) *Session {
	return &Session{
		oper:    oper,
		connect: connFunc,
	}
}

func init() {
	// init all support database
	initSessM["mysql"] = newSess(&mysqlOper{}, mysqlConnect)
}

var (
	// global session
	sess *Session
	mu   sync.Mutex
)

// Init will take out the connection configuration according
// to "key", select the database type according to "dbType"
// (if the connection configuration does not exist or the
// database type is not supported, an error will be returned),
// and then connect to the database and assign values to the
// global session.
//
// If the database connection is wrong, the function will also
// return an error. If Init is called repeatedly, the global
// session will be overwritten.
//
// Through Get(), you can get the global session and call its
// methods. If Init is not called when Get is called, the program
// will exit abnormally. You can use MustInit() to check whether
// Init() is called and successful.
func Init(key, dbType string) error {
	sess = initSessM[dbType]
	if sess == nil {
		return fmt.Errorf(
			"unsupport database type: \"%s\"", dbType)
	}

	cfg, err := conn.Get(key)
	if err != nil {
		return errors.Trace("read connection", err)
	}

	sess.cfg = cfg
	sess.db, err = sess.connect(cfg)
	if err != nil {
		return errors.Trace("init database", err)
	}
	sess.oper.Init(sess)

	return nil
}

// MustInit checks whether Init is called and initialized normally.
func MustInit() error {
	if sess != nil {
		return nil
	}
	return errors.New(`need database connection, ` +
		`did you forget to add "--conn" flag?`)
}

// Get returns the global session. If Init is not called before
// the call, the program will exit abnormally.
func Get() *Session {
	if sess == nil {
		log.Fatal("database is not init!")
	}
	return sess
}

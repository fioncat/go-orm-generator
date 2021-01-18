package rdb

import (
	"time"

	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/store"
)

var (
	// Indicates whether to enable data table caching.
	// If true, the data table will be cached on the
	// local disk. Repeated reading of the data table
	// will be read locally. This can reduce the interaction
	// with the remote database to speed up the calling
	// speed of Desc.
	// The default is closed, if you want to open, you need
	// to manually modify the variable value.
	ENABLE_TABLE_CACHE = false

	// The lifetime of the table cache. If the table cache
	// duration is greater than this value, it will be
	// invalidated and Desc will retrieve data from the
	// database again.
	TABLE_CACHE_TTL = time.Hour * 24
)

// CacheTable represents the structure of the data table
// cached on the disk. It implements the Table interface,
// and any Table interface can be converted to it.
type CacheTable struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`

	Fields map[string]*CacheField `json:"fields"`
}

func (t *CacheTable) GetName() string         { return t.Name }
func (t *CacheTable) GetComment() string      { return t.Comment }
func (t *CacheTable) Field(name string) Field { return t.Fields[name] }

func (t *CacheTable) FieldNames() []string {
	names := make([]string, 0, len(t.Fields))
	for name := range t.Fields {
		names = append(names, name)
	}
	return names
}

// convert Table interface into CacheTable struct.
func (t *CacheTable) fromInter(it Table) {
	t.Name = it.GetName()
	t.Comment = it.GetComment()

	fieldNames := it.FieldNames()
	t.Fields = make(map[string]*CacheField, len(fieldNames))
	for _, fieldName := range fieldNames {
		field := it.Field(fieldName)
		if field == nil {
			continue
		}
		t.Fields[fieldName] = &CacheField{
			Name:      field.GetName(),
			Comment:   field.GetComment(),
			Type:      field.GetType(),
			IsPrimary: field.IsPrimaryKey(),
			AutoIncr:  field.IsAutoIncr(),
		}
	}
}

// CacheField represents the structure of the data table field
// cached on the disk. It implements the Field interface,
// and any Table interface can be converted to it.
type CacheField struct {
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	Type      string `json:"type"`
	IsPrimary bool   `json:"is_primary"`
	AutoIncr  bool   `json:"auto_incr"`
}

func (f *CacheField) GetName() string    { return f.Name }
func (f *CacheField) GetComment() string { return f.Comment }
func (f *CacheField) GetType() string    { return f.Type }
func (f *CacheField) IsPrimaryKey() bool { return f.IsPrimary }
func (f *CacheField) IsAutoIncr() bool   { return f.AutoIncr }

// get table from the local disk, if cache miss, returns nil
func getCacheTable(key string) Table {
	var cacheTable CacheTable
	ok, err := store.Load(key, &cacheTable)
	if err != nil {
		log.Errorf("read table cache failed: key=%s, err=%v",
			key, err)
		return nil
	}
	if !ok {
		return nil
	}

	log.Infof("table cache hit: %s", key)
	return &cacheTable
}

// save table to local disk. If an error occurs, it
// will not return, but will log.
func saveCacheTable(key string, table Table) {
	cacheTable := new(CacheTable)
	cacheTable.fromInter(table)

	err := store.Save(key, cacheTable, TABLE_CACHE_TTL)
	if err != nil {
		log.Errorf("save cache failed: key=%s, err=%v",
			key, err)
	}
	log.Infof("save cache success: %s", key)
}

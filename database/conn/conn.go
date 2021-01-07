package conn

import (
	"fmt"

	"github.com/fioncat/go-gendb/store"
)

// Config represents the general database connection
// configuration. It stores the information needed to
// connect to the database, some of which can be empty.
// This structure will be used by different databases to
// create connections.
type Config struct {

	// Key is used to uniquely locate a certain
	// connection configuration.
	Key string `json:"key"`

	// Database connection basic information, address,
	// user and password (used for authentication, can
	// be empty)
	Addr     string `json:"addr"`
	User     string `json:"user"`
	Password string `json:"password"`

	// database name, can be empty
	Database string `json:"database"`
}

// key prefix store in the disk.
const keyPrefix = "conn_"

// Set save the database connection configuration "cfg"
// to the local disk. The key is used to retrieve this
// configuration through Get. This function will trigger
// a file IO operation.
// If the configuration corresponding to the key already
// exists on the disk, this function will replace it with
// the new configuration.
func Set(key string, cfg *Config) error {
	cfg.Key = key
	key = keyPrefix + key
	return store.Save(key, cfg, 0)
}

// Get obtains the database configuration of the specified
// key from the disk. The configuration is set by Set. If
// the configuration does not exist, an error will be returned.
// This function will trigger a file IO operation.
func Get(key string) (*Config, error) {
	oriKey := key
	key = keyPrefix + key

	cfg := new(Config)
	ok, err := store.Load(key, cfg)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf(`can not `+
			`find connection "%s"`, oriKey)
	}

	return cfg, nil
}

// Remove will delete the configuration corresponding to
// the key from the disk.
func Remove(key string) error {
	key = keyPrefix + key
	return store.Remove(key)
}

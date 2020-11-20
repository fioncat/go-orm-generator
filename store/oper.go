package store

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
)

type Item struct {
	Data   []byte
	Expire int64
}

func Save(base, name string, o interface{}, ttl time.Duration) error {
	data, err := json.Marshal(o)
	if err != nil {
		return errors.Trace("marshal failed", err)
	}

	var item Item
	item.Data = data
	if ttl > 0 {
		item.Expire = time.Now().Add(ttl).Unix()
	}

	data, err = json.Marshal(item)
	if err != nil {
		return errors.Trace("marshal failed", err)
	}

	path := filepath.Join(baseHome(), base, name)
	dir := filepath.Dir(path)
	os.MkdirAll(dir, os.ModePerm)
	return ioutil.WriteFile(path, data, 0644)
}

func Get(base, name string, v interface{}) (bool, error) {
	path := filepath.Join(baseHome(), base, name)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if err == os.ErrNotExist {
			return false, nil
		}
		return false, err
	}

	var item Item
	err = json.Unmarshal(data, &item)
	if err != nil {
		return false, errors.Trace("unmarshal failed", err)
	}

	if item.Expire > 0 {
		if time.Now().Unix() >= item.Expire {
			err = Remove(base, name)
			if err != nil {
				log.Errorf("remove expired file"+
					"%s/%s failed: %v", base, name, err)
			}
			return false, nil
		}
	}

	err = json.Unmarshal(item.Data, v)
	if err != nil {
		return false, errors.Trace("unmarshal failed", err)
	}
	return true, nil
}

func Remove(base, name string) error {
	path := filepath.Join(baseHome(), base, name)
	return os.Remove(path)
}

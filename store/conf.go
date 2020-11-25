package store

import (
	"encoding/json"
	"io/ioutil"
)

func UnmarshalConf(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

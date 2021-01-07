package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/fioncat/go-gendb/misc/errors"
)

// encode data to payload
func encodePayload(data []byte, expire int64) []byte {
	dataLen := uint32(len(data))

	var intBuf = make([]byte, 8)
	binary.BigEndian.PutUint32(intBuf, dataLen)

	if expire <= 0 {
		payload := make([]byte, 0, dataLen+4)
		payload = append(payload, intBuf[:4]...)
		payload = append(payload, data...)
		return payload
	}

	// 4 Bytes datalen, 8 Bytes expire date
	payload := make([]byte, 0, dataLen+12)
	payload = append(payload, intBuf[:4]...)
	payload = append(payload, data...)
	binary.BigEndian.PutUint64(intBuf, uint64(expire))
	payload = append(payload, intBuf...)

	return payload
}

// decode payload to data
func decodePayload(payload []byte) ([]byte, bool) {
	reader := bytes.NewReader(payload)
	// First 4 bytes is datalen
	dataLenData := make([]byte, 4)
	_, err := reader.Read(dataLenData)
	if err == io.EOF {
		return nil, false
	}

	dataLen := int(binary.BigEndian.Uint32(dataLenData))
	if dataLen <= 0 {
		return nil, false
	}

	data := make([]byte, dataLen)
	_, err = reader.Read(data)
	if err == io.EOF {
		return nil, false
	}

	if reader.Len() == 0 {
		// reach the end, the payload has not expire date
		return data, true
	}

	// expire data
	expireData := make([]byte, 8)
	_, err = reader.Read(expireData)
	if err == io.EOF {
		return nil, false
	}

	expire := int64(binary.BigEndian.Uint64(expireData))
	if expire <= 0 {
		return nil, false
	}

	// check if the data is expired
	if time.Now().Unix() >= expire {
		return nil, true
	}

	return data, true
}

// Save saves the object to disk. The specific path is
// the "basePath" directory under the home directory
// (it will be created automatically).
//
// The object will be serialized and stored in json format.
// ttl represents the survival time of the object, and it
// will be stored together with the json data. When calling
// Load, ttl will be checked. See the Load document for details.
//
// name is used to retrieve the object from the disk. If the
// object corresponding to name already exists when Save is
// called, it will be aligned and overwritten.
func Save(name string, v interface{}, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return errors.Trace("marshal", err)
	}

	var expire int64
	if ttl > 0 {
		expire = time.Now().Add(ttl).Unix()
	}

	payload := encodePayload(data, expire)
	path := filepath.Join(baseHome(), name)

	return ioutil.WriteFile(path, payload, 0644)
}

// Load will load the data of the specified name from the
// disk, and then deserialize the data to "v" and return.
// The process will check ttl, if the object expires, it
// will return "false, nil", and call Remove to delete the
// object. If the object does not exist, it will also return
// "false, nil". If the object exists and the IO and deserialization
// does not produce an error, it will return "true, nil".
func Load(name string, v interface{}) (bool, error) {
	path := filepath.Join(baseHome(), name)
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsExist(err) {
			// file not exists
			return false, nil
		}
		return false, errors.Trace(
			"get file state", err)
	}

	payload, err := ioutil.ReadFile(path)
	if err != nil {
		return false, errors.Trace(
			"read file", err)
	}

	data, ok := decodePayload(payload)
	if !ok {
		return false, errors.New(
			"%s: data bad format", path)
	}
	if len(data) == 0 {
		// file is expired, remove it
		err = Remove(name)
		if err != nil {
			return false, errors.Trace(
				"remove expire file", err)
		}
		return false, nil
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return false, errors.Trace("unmarshal", err)
	}

	return true, nil
}

// Remove delete object in the disk.
func Remove(name string) error {
	path := filepath.Join(baseHome(), name)
	return os.Remove(path)
}

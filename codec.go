package goutil

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

/// gob

func GobEncode(obj interface{}) (*bytes.Buffer, error) {
	cached := &bytes.Buffer{}
	enc := gob.NewEncoder(cached)
	if err := enc.Encode(obj); err != nil {
		return nil, err
	} else {
		return cached, nil
	}
}

func GobDecode(data []byte, obj interface{}) error {
	cached := bytes.NewBuffer(data)
	dec := gob.NewDecoder(cached)
	return dec.Decode(obj)
}

/// json

func JsonEncode(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func JsonDecode(jdata []byte, obj interface{}) error {
	err := json.Unmarshal(jdata, obj)
	if err != nil {
		return err
	}
	return nil
}

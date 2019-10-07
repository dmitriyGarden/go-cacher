package cacher

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"io/ioutil"
)

type storageData struct {
	Value      string       `json:"v"`
	Dependency []Dependency `json:"d"`
}

func marshalData(value string, dependency ...IDependency) ([]byte, error) {
	d := storageData{
		Value:      value,
		Dependency: make([]Dependency, len(dependency)),
	}
	for i := range dependency {
		d.Dependency[i] = Dependency{
			Key:   dependency[i].GetKey(),
			Value: dependency[i].GetValue(),
		}
	}
	return json.Marshal(&d)
}

func unmarshalData(data []byte) (*storageData, error) {
	d := new(storageData)
	err := json.Unmarshal(data, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func compressData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	zw, err := flate.NewWriter(&b, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	_, err = zw.Write(data)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decompressData(data []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewBuffer(data))
	defer r.Close()
	return ioutil.ReadAll(r)
}

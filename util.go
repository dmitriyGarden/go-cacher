package cacher

import (
	"bytes"
	"encoding/json"
	"github.com/andrew-d/lzma"
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
	w := lzma.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decompressData(data []byte) ([]byte, error) {
	r := lzma.NewReader(bytes.NewBuffer(data))
	defer r.Close()
	return ioutil.ReadAll(r)
}

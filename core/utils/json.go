package utils

import (
	"bytes"
)

/*
	把一个对象转换成map
*/
func ChangeMap(v interface{}) (map[string]interface{}, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	// err = json.Unmarshal(bs, &m)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

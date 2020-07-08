package protocol

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func (m *MapInt32Int32) FromDB(data []byte) (err error) {
	err = m.UnmarshalJSON(data)
	return
}

func (m *MapInt32Int32) ToDB() (data []byte, err error) {
	data, err = m.MarshalJSON()
	return
}

func (m *MapInt32Int32) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	tmp := map[int32]int32{}
	m.Range(func(key, value int32) bool {
		tmp[key] = value
		return true
	})
	ret, err := json.Marshal(&tmp)
	if err != nil {
		return nil, err
	}
	return ret, nil

}

func (m *MapInt32Int32) UnmarshalJSON(b []byte) error {
	if m == nil {
		return errors.New(" Unmarshal(non-pointer MapInt32Int32)")
	}
	tmp := map[int32]int32{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	for k, v := range tmp {
		m.Store(k, v)
	}
	return nil
}

func (m *MapInt32Int32) String() string {
	if m == nil {
		return "{}"
	}
	builder := strings.Builder{}
	builder.WriteString("{")
	m.Range(func(key, value int32) bool {
		builder.WriteString("{")
		builder.WriteString(strconv.FormatInt(int64(key), 10))
		builder.WriteString(":")
		builder.WriteString(strconv.FormatInt(int64(value), 10))
		builder.WriteString("}")
		builder.WriteString(",")
		return true
	})
	builder.WriteString("}")
	return builder.String()
}

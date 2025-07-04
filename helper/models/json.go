package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type JsonScan[T any] struct {
	v *T
}

func NewJsonScan[T any](v T) JsonScan[T] {
	return JsonScan[T]{v: &v}
}

func parseJsonObject(data []byte) (map[string]interface{}, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}
	return temp, nil
}

func parseJsonSlice(data []byte) ([]interface{}, error) {
	var temp []interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}
	return temp, nil
}

func (j *JsonScan[T]) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	j.v = new(T)
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			return nil
		}
		if !json.Valid(v) {
			return fmt.Errorf("json format was invalid")
		}

		var jsdata interface{}

		mdata, merr := parseJsonObject(v)
		if merr != nil && !strings.Contains(merr.Error(), "cannot unmarshal array into Go value of type") {
			return merr
		}
		if mdata != nil {
			jsdata = mdata
		}

		sdata, serr := parseJsonSlice(v)
		if serr != nil && !strings.Contains(serr.Error(), "cannot unmarshal object into Go value of type") {
			return serr
		}
		if sdata != nil {
			jsdata = sdata
		}

		// Marshal back into []byte
		data, err := json.Marshal(jsdata)
		if err != nil {
			return err
		}

		// Assign the JSON data to TenantConfig
		if err := json.Unmarshal(data, j.v); err != nil {
			return err
		}
	}

	return nil
}

func (j JsonScan[T]) Value() (driver.Value, error) {
	if j.v == nil {
		return nil, nil
	}

	return json.Marshal(j.v)
}

func (j JsonScan[T]) MarshalJSON() ([]byte, error) {
	if reflect.TypeOf(j.v).Elem().Kind() == reflect.Slice {
		if reflect.ValueOf(j.v).IsNil() && reflect.ValueOf(j.v).IsZero() {
			return json.Marshal([]interface{}{})
		}
	}

	return json.Marshal(j.v)
}

func (j *JsonScan[T]) UnmarshalJSON(src []byte) error {
	if len(src) == 0 {
		return nil
	}

	if j.v == nil {
		j.v = new(T)
	}
	return json.Unmarshal(src, j.v)
}

func (j *JsonScan[T]) Data() T {
	if reflect.ValueOf(j.v).IsNil() || reflect.ValueOf(j.v).IsZero() {
		j.v = new(T)
	}
	return *j.v
}

func (j *JsonScan[T]) Set(v T) *JsonScan[T] {
	j.v = &v
	return j
}

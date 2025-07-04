package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/spf13/cast"
)

type EnumScan[T ~string] struct {
	v *T
}

func NewEnumScan[T ~string](value string) EnumScan[T] {
	v := T(value)
	return EnumScan[T]{v: &v}
}

func (e *EnumScan[T]) Scan(value interface{}) error {
	if value == nil {
		var zero T
		e.v = &zero
		return nil
	}

	data := T(cast.ToString(value))
	e.v = &data
	return nil
}

func (j EnumScan[T]) Value() (driver.Value, error) {
	if j.v == nil {
		return nil, nil
	}

	return string(*j.v), nil
}

func (j EnumScan[T]) MarshalJSON() ([]byte, error) {
	if j.v == nil {
		return []byte(""), nil
	}
	return json.Marshal(j.v)
}

func (j *EnumScan[T]) UnmarshalJSON(src []byte) error {
	if len(src) == 0 {
		return nil
	}

	if j.v == nil {
		j.v = new(T)
	}
	return json.Unmarshal(src, j.v)
}

func (j *EnumScan[T]) Data() T {
	if j.v == nil {
		var zero T
		return zero
	}

	return *j.v
}

func (j *EnumScan[T]) Set(v T) {
	j.v = &v
}

func (j EnumScan[T]) String() string {
	if j.v == nil {
		return ""
	}
	return string(*j.v)
}

func (j EnumScan[T]) IsZero() bool {
	return j.v == nil || *j.v == ""
}

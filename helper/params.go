package helper

import (
	"github.com/gofrs/uuid"
	"github.com/guregu/null/zero"
	"reflect"
)

func GetInt64WithParams(value interface{}) int64 {
	var num int64
	if value == nil {
		return 0
	}

	if reflect.TypeOf(value).Kind() == reflect.Float64 {
		num = int64(value.(float64))
	} else if reflect.TypeOf(value).Kind() == reflect.Int64 {
		num = value.(int64)
	}

	return num
}

func GetStringWithParams(value interface{}) string {
	var text string
	if value == nil {
		return ""
	}

	if reflect.TypeOf(value).Kind() == reflect.String {
		text = value.(string)
	}

	return text
}

func GetFloat64WithParams(value interface{}) float64 {
	var num float64
	if value == nil {
		return 0
	}

	if reflect.TypeOf(value).Kind() == reflect.Float64 {
		num = value.(float64)
	} else if reflect.TypeOf(value).Kind() == reflect.Int64 {
		num = float64(value.(int64))
	}

	return num
}

func GetUUIDWithParams(value interface{}) *uuid.UUID {
	if value == nil {
		return nil
	}
	uid, err := uuid.FromString(value.(string))
	if err != nil {
		return nil
	}
	return &uid
}

func GetZeroIntWithParams(value interface{}) zero.Int {
	var num zero.Int
	if value == nil {
		return num
	}

	if reflect.TypeOf(value).Kind() == reflect.Int64 {
		num = zero.IntFrom(value.(int64))
	} else if reflect.TypeOf(value).Kind() == reflect.Float64 {
		num = zero.IntFrom(int64(value.(float64)))
	}

	return num
}

func GetZeroStringWithParams(value interface{}) zero.String {
	var text zero.String
	if value == nil {
		return text
	}

	if reflect.TypeOf(value).Kind() == reflect.String {
		text = zero.StringFrom(value.(string))
	}

	return text
}

func GetZeroFloatWithParams(value interface{}) zero.Float {
	var num zero.Float
	if value == nil {
		return num
	}

	if reflect.TypeOf(value).Kind() == reflect.Float64 {
		num = zero.FloatFrom(value.(float64))
	} else if reflect.TypeOf(value).Kind() == reflect.Int64 {
		num = zero.FloatFrom(float64(value.(int64)))
	}

	return num
}
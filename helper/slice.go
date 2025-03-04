package helper

import (
	"reflect"
)

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func MergeMap(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	for key, value := range m2 {
		m1[key] = value
	}
	return m1
}

func MapKeyToArray(m map[string]interface{}) []interface{} {
	var keysArray []interface{}
	for key, _ := range m {
		keysArray = append(keysArray, key)
	}
	return keysArray
}

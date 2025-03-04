package helper

import (
	"github.com/fatih/structs"
	"strconv"
	"strings"
)

func GetValueFromTag(model interface{}, field string, tag string) string {
	var tagVal string

	m := structs.New(model)
	if f, ok := m.FieldOk(field); ok {
		tagVal = f.Tag(tag)
	}

	return tagVal
}

func SetStringIFNil(value interface{}) string {
	if value == nil {
		return ""
	}
	return value.(string)
}

func SetFloat64IFNil(value interface{}) float64 {
	floatValue, _ := strconv.ParseFloat(value.(string), 64)
	if value == nil {
		return floatValue
	}
	return floatValue
}

func SetInt32IFNil(value interface{}) int32 {
	var intvalue32 int32
	intValue64, _ := strconv.ParseInt(value.(string), 10, 32)
	intvalue32 = int32(intValue64)
	if value == nil {
		return intvalue32
	}
	return intvalue32
}

func SetBoolIFNil(value interface{}) bool {
	if value == nil {
		return false
	}
	return value.(bool)
}

func IndexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

func GetSlug(topic string) string {
	splitBy := " "
	replaceBy := "-"
	return strings.ReplaceAll(topic, splitBy, replaceBy)
}

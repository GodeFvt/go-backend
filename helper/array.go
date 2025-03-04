package helper

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func EscapeCharacter(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\a", "")
	s = strings.ReplaceAll(s, "\b", "")
	s = strings.ReplaceAll(s, "\f", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\v", "")
	s = strings.ReplaceAll(s, `\u`, "")
	return s
}

func GetOneQueryParameterFromURL(urlLink string, queryVar string) (string, error) {
	var val string
	u, err := url.Parse(urlLink)
	if err != nil {
		return "", err
	}
	q := u.Query()
	val = q.Get(queryVar)

	return val, nil
}

func ContainStringWithString(str1 string, str2 string) bool {
	return str1 == str2
}

func MapContainStringWithString(str1 string, strs []string) bool {
	for _, str := range strs {
		if ContainStringWithString(str1, str) {
			return true
		}
	}
	return false

}

func ToTileWithOutRomanNumber(str string) string {
	if str == "I" || str == "II" || str == "III" || str == "IV" || str == "V" {
		return str
	} else if str == "AND" || str == "TO" || str == "OR" || str == "OF" {
		return strings.ToLower(str)
	}
	return strings.Title(strings.ToLower(str))
}

func GetListConfigKey(prefixKey string, amount int) []string {
	var keys = []string{}
	if amount > 0 {
		for i := 0; i < amount; i++ {
			numStr := strconv.Itoa(i + 1)
			key := fmt.Sprintf("%s_%s", prefixKey, numStr)
			keys = append(keys, key)
		}
	}
	return keys
}

func RemoveIndex(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}

func SplitStringTwoArray(text string, seperrate string) (string, string) {
	var texts = strings.Split(text, seperrate)
	return texts[0], texts[1]
}

func SplitStringArray(text string, seperrate string) []string {
	var texts = strings.Split(text, seperrate)
	return texts
}


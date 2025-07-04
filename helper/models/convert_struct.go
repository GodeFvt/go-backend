package models

import "encoding/json"

func ConvertStruct[Src any, Dst any](src Src) (Dst, error) {
	var dst Dst

	// Marshal src to JSON
	bytes, err := json.Marshal(src)
	if err != nil {
		return dst, err
	}

	// Unmarshal into dst
	err = json.Unmarshal(bytes, &dst)
	return dst, err
}

func CopyJSON(src any, dst any) error {
	// Marshal src to JSON
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	// Unmarshal into dst
	return json.Unmarshal(bytes, &dst)
}

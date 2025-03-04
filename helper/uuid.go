package helper

import (
	"reflect"

	"github.com/globalsign/mgo/bson"
	"github.com/gofrs/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var uuidBinaryKind = bson.BinaryUUID
var bsonUUIDType = reflect.TypeOf(primitive.Binary{})

func ConvertToUUIDAndBinary(v interface{}) (*uuid.UUID, *bson.Binary) {
	var uid *uuid.UUID
	var key *bson.Binary
	if v != nil {
		if reflect.TypeOf(v).Kind() == reflect.String {
			if v.(string) != "" {
				u := uuid.FromStringOrNil(v.(string))
				uid = &u
				binary := SetUUIDBson(u)
				key = binary
			}
		} else if reflect.TypeOf(v).Kind() == reflect.TypeOf(uuid.UUID{}).Kind() {
			if v.(uuid.UUID) != (uuid.UUID{}) {
				u := v.(uuid.UUID)
				uid = &u
				binary := SetUUIDBson(u)
				key = binary
			}
		} else if reflect.ValueOf(v).Type() == reflect.ValueOf(bson.Binary{}).Type() {
			binary := v.(bson.Binary)
			key = &binary
			u := GetUUIDFromBson(&binary)
			uid = &u
		} else if reflect.TypeOf(v) == bsonUUIDType {
			id := uuid.FromBytesOrNil(v.(primitive.Binary).Data)
			uid = &id
		}
	}
	if uid != nil {
		if uid.String() == "00000000-0000-0000-0000-000000000000" {
			uid = nil
		}
	}
	return uid, key
}

func SetUUIDBson(id uuid.UUID) *bson.Binary {
	return &bson.Binary{
		Kind: uuidBinaryKind,
		Data: id.Bytes(),
	}
}

func GetUUIDFromBson(binary *bson.Binary) uuid.UUID {
	return uuid.FromBytesOrNil(binary.Data)
}

func ToUUIDBson(v interface{}) *bson.Binary {
	if v == nil {
		return nil
	}
	t := reflect.TypeOf(v).String()
	switch t {
	case "string":
		uid := uuid.FromStringOrNil(v.(string))
		return SetUUIDBson(uid)
	case "uuid.UUID":
		return SetUUIDBson(v.(uuid.UUID))
	case "*uuid.UUID":
		uid := v.(*uuid.UUID)
		return SetUUIDBson(*uid)
	}

	return nil
}

func GetBsonSlice(uuids []*uuid.UUID) []*bson.Binary {
	var bs = make([]*bson.Binary, 0)
	if uuids != nil {
		if len(uuids) > 0 {
			for _, uid := range uuids {
				key := SetUUIDBson(*uid)
				bs = append(bs, key)
			}
		}
	}
	return bs
}

func UUIDToSliceString(slice []*uuid.UUID) []string {
	var uids = make([]string, 0)
	if slice != nil {
		if len(slice) > 0 {
			for _, uid := range slice {
				if uid != nil {
					uids = append(uids, uid.String())
				}
			}
		}
	}
	return uids
}

func FindInSliceUUID(slice []*uuid.UUID, uid *uuid.UUID) (exists bool, index int) {
	if slice != nil {
		if len(slice) > 0 {
			for indexItem, item := range slice {
				if item.String() == uid.String() {
					exists = true
					index = indexItem
					break
				}
			}
		}
	}
	return
}

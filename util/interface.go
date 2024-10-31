package util

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/evolutionlandorg/evo-backend/util/log"
)

func ToString(i interface{}) string {
	var val string
	switch i := i.(type) {
	case string:
		val = i
	case []byte:
		val = string(i)
	default:
		b, _ := json.Marshal(i)
		val = string(b)
	}
	return val
}

func UnmarshalAny(r interface{}, raw interface{}) {
	switch raw := raw.(type) {
	case string:
		_ = json.Unmarshal([]byte(raw), &r)
	case []uint8:
		_ = json.Unmarshal(raw, &r)
	default:
		b, _ := json.Marshal(raw)
		_ = json.Unmarshal(b, r)
	}
}

func Debug(i interface{}) {
	var val string
	switch i := i.(type) {
	case string:
		val = i
	case []byte:
		val = string(i)
	case error:
		val = i.Error()
	default:
		b, _ := json.MarshalIndent(i, "", "  ")
		val = string(b)
	}
	log.Debug(val)
}

func GetFieldValByTag(tag, key string, s interface{}) (result interface{}, err error) {
	rt := reflect.ValueOf(s).Elem()
	if rt.Kind() != reflect.Struct {
		return nil, errors.New("bad type")
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		t := rt.Type().Field(i)
		if t.Type.Kind() == reflect.Struct && t.Type.Name() != "Decimal" && t.Type.Name() != "Time" {
			if result, _ = GetFieldValByTag(tag, key, f.Addr().Interface()); result != nil {
				return result, nil
			}
		} else if t.Type.Kind() == reflect.Ptr {
			if result, _ = GetFieldValByTag(tag, key, f.Interface()); result != nil {
				return result, nil
			}
		} else {
			v := strings.Split(t.Tag.Get(key), ",")[0] // use split to ignore tag "options"
			if v == tag {
				return f.Interface(), nil
			}
		}

	}
	return nil, nil
}

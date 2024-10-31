package util

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

func StructToSql(src interface{}) (wheres []string, values []interface{}) {
	elem := reflect.TypeOf(src)
	value := reflect.ValueOf(src)
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		v := value.FieldByName(f.Name)
		if v.IsZero() {
			continue
		}
		tableName := f.Tag.Get("table_name")
		symbol := f.Tag.Get("symbol")
		if symbol == "" {
			symbol = "="
		}
		name := f.Tag.Get("json")
		switch symbol {
		case "IN", "in":
			wheres = append(wheres, fmt.Sprintf("%s.%s IN (?)", tableName, name))
			values = append(values, RemoveEmptyStrings(strings.Split(cast.ToString(v.Interface()), ",")))
			continue
		case "=":
			wheres = append(wheres, fmt.Sprintf("%s.%s = ?", tableName, name))
		case ">=":
			wheres = append(wheres, fmt.Sprintf("%s.%s >= ?", tableName, name))
		case "<=":
			wheres = append(wheres, fmt.Sprintf("%s.%s <= ?", tableName, name))
		default:
			log.Panicf("not support symbol %s", symbol)
		}
		values = append(values, v.Interface())
	}
	return
}

package sqlhelpers

import (
	"fmt"
	"reflect"
	"strings"
)

func GetArgs(val interface{}, quote bool, skipZero bool, skipFields ...string) (fields []string, args []interface{}, err error) {
	to := reflect.TypeOf(val)
	vo := reflect.ValueOf(val)
	if to.Kind() == reflect.Ptr {
		to = to.Elem()
		vo = vo.Elem()
	}
	if to.Kind() != reflect.Struct {
		return nil, nil, NotAStruct
	}

	var f reflect.StructField
	var v reflect.Value
	for i := 0; i < to.NumField(); i++ {
		f = to.Field(i)
		v = vo.Field(i)
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if Contains(skipFields, name) {
			continue
		}
		if v.CanInterface() && (!skipZero || !v.IsZero()) {
			if quote {
				fields = append(fields, `"`+name+`"`)
			} else {
				fields = append(fields, name)
			}
			args = append(args, v.Interface())
		}
		// args = append(args, v.)
	}
	return fields, args, nil
}

func CreateTable(typ interface{}, name string) string {
	names := Tags(typ, "json")
	props := Tags(typ, "db")
	fields := []string{}
	for i, name := range names {
		name = strings.Split(name, ",")[0]
		if props[i] != "" {
			t := strings.Split(props[i], ":")
			if t[0] == "field" {
				fields = append(fields, `"`+name+`"`+" "+t[1])
			}
		}
	}

	return fmt.Sprintf(`create table if not exists %v (%v)`, name, strings.Join(fields, ",\n"))
}

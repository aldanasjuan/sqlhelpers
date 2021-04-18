package sqlhelpers

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func NowMiliseconds() int {
	return int(time.Now().UnixNano() / 1000000)
}

func Fields(v interface{}, tag string) *Set {
	tags := Map(Tags(v, tag), func(s string, i int) string {

		return strings.Split(s, ",")[0]
	})

	set := make(Set)
	for _, s := range tags {
		set.Add(s)
	}
	return &set
}
func Args(l int) string {
	b := strings.Builder{}
	b.Grow(l * 3)
	for i := 0; i < l; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "$%d", i+1)
	}
	return b.String()
}

func Map(original []string, callback func(string, int) string) []string {
	res := make([]string, 0, len(original))
	for i, v := range original {
		res = append(res, callback(v, i))
	}
	return res
}

func Tags(v interface{}, tag string) (res []string) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		v := f.Tag.Get(tag)
		res = append(res, v)
	}

	return res
}

func Contains(vals []string, s string) bool {
	for _, val := range vals {
		if s == val {
			return true
		}
	}
	return false
}

func Reset(val interface{}) error {
	vo := reflect.ValueOf(val)
	if vo.Kind() != reflect.Ptr {
		return NotAStruct
	}
	vo = vo.Elem()
	if vo.Kind() != reflect.Struct {
		return NotAStruct
	}

	for i := 0; i < vo.NumField(); i++ {
		f := vo.Field(i)
		if f.CanSet() {
			t := f.Type()
			f.Set(reflect.Zero(t))

		}
	}
	return nil
}

func StructMap(v interface{}) (Table, error) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, NotAStruct
	}
	res := Table{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Name
		res[name] = Field{
			JSON: strings.Split(field.Tag.Get("json"), ",")[0],
			DB:   field.Tag.Get("db"),
		}
	}
	return res, nil
}

type Table map[string]Field

type Field struct {
	JSON string `json:"json,omitempty"`
	DB   string `json:"db,omitempty"`
}

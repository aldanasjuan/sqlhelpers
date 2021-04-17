package sqlhelpers

import (
	"bytes"
	"fmt"
)

type Filter struct {
	Field    string      `json:"f,omitempty"`
	Operator string      `json:"o,omitempty"`
	Value    interface{} `json:"v,omitempty"`
	Type     string      `json:"t,omitempty"`
	Filters  Filters     `json:"fs,omitempty"`
}

type Filters []Filter

func (f Filters) Values(start int, and bool, operators *Set, fields *Set) (string, []interface{}) {
	b := bytes.Buffer{}
	args := []interface{}{}
	l := start
	for _, filter := range f {
		switch filter.Type {
		case "or":
			inner, a := filter.Filters.Values(l, false, operators, fields)
			if len(a) > 0 {
				args = append(args, a...)
				fmt.Fprintf(&b, `%v (%v) `, And(and), inner)
				l += len(a)
			}
		case "and":
			inner, a := filter.Filters.Values(l, true, operators, fields)
			if len(a) > 0 {
				args = append(args, a...)
				fmt.Fprintf(&b, `%v (%v) `, And(and), inner)
				l += len(a)
			}
		default:
			if operators.Exists(filter.Operator) && fields.Exists(filter.Field) {
				if len(args) > 0 {
					b.WriteString(And(and))
				}
				args = append(args, filter.Value)
				fmt.Fprintf(&b, `%q %v $%d `, filter.Field, filter.Operator, l)
				l++
			}
		}
	}
	return b.String(), args
}

//And returns "and" if true or "or" if false
func And(b bool) string {
	if b {
		return "and "
	}
	return "or "
}

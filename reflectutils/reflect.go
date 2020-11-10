package reflectutils

import (
	"reflect"
	"strings"
)

func GetStructName(value interface{}) string {
	typ := reflect.TypeOf(value).String()
	sub := "."
	tSlice := strings.Split(typ, sub)
	name := tSlice[len(tSlice)-1]
	return name
}

package convertutils

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/Yamiyo/common/numberutils"
	"github.com/Yamiyo/common/rlp"
)

var (
	timeTime = reflect.TypeOf(time.Time{})
)

func DeepFields(ifaceType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	for i := 0; i < ifaceType.NumField(); i++ {
		v := ifaceType.Field(i)
		if v.Anonymous && v.Type.Kind() == reflect.Struct {
			fields = append(fields, DeepFields(v.Type)...)
		} else {
			fields = append(fields, v)
		}
	}

	return fields
}

func StructCopy(DstStructPtr interface{}, SrcStructPtr interface{}) {
	srcv := reflect.ValueOf(SrcStructPtr)
	dstv := reflect.ValueOf(DstStructPtr)
	srct := reflect.TypeOf(SrcStructPtr)
	dstt := reflect.TypeOf(DstStructPtr)
	if srct.Kind() != reflect.Ptr || dstt.Kind() != reflect.Ptr ||
		srct.Elem().Kind() == reflect.Ptr || dstt.Elem().Kind() == reflect.Ptr {
		panic("Fatal error:type of parameters must be Ptr of value")
	}
	if srcv.IsNil() || dstv.IsNil() {
		panic("Fatal error:value of parameters should not be nil")
	}
	srcV := srcv.Elem()
	dstV := dstv.Elem()
	srcfields := DeepFields(reflect.ValueOf(SrcStructPtr).Elem().Type())
	for _, v := range srcfields {
		if v.Anonymous {
			continue
		}
		dst := dstV.FieldByName(v.Name)
		src := srcV.FieldByName(v.Name)
		if !dst.IsValid() {
			continue
		}
		if src.Kind() == reflect.String && dst.Type() == reflect.TypeOf(sql.NullString{}) && dst.CanSet() {
			nullString := &sql.NullString{}
			nullString.Scan(src.String())
			dst.Set(reflect.ValueOf(*nullString))
			continue
		}
		if src.Type() == reflect.TypeOf(sql.NullString{}) && dst.Kind() == reflect.String && dst.CanSet() {
			obj := src.Interface()
			nullString := obj.(sql.NullString)
			s := nullString.String
			dst.Set(reflect.ValueOf(s))
			continue
		}
		if src.Kind() >= reflect.Int && src.Kind() <= reflect.Int64 && dst.Type() == reflect.TypeOf(sql.NullInt64{}) && dst.CanSet() {
			nullInt64 := &sql.NullInt64{}
			i := src.Int()
			nullInt64.Scan(int64(i))
			dst.Set(reflect.ValueOf(*nullInt64))
			continue
		}
		if src.Type() == reflect.TypeOf(sql.NullInt64{}) && dst.Kind() == reflect.Int64 && dst.CanSet() {
			obj := src.Interface()
			nullInt64 := obj.(sql.NullInt64)
			i := nullInt64.Int64
			dst.Set(reflect.ValueOf(i))
			continue
		}
		if src.Type() == dst.Type() && dst.CanSet() {
			dst.Set(src)
			continue
		}
		if src.Kind() == reflect.Ptr && !src.IsNil() && src.Type().Elem() == dst.Type() {
			dst.Set(src.Elem())
			continue
		}
		if dst.Kind() == reflect.Ptr && dst.Type().Elem() == src.Type() {
			dst.Set(reflect.New(src.Type()))
			dst.Elem().Set(src)
			continue
		}
	}
	return
}

func FloatRound(structPtr interface{}) {
	floatRoundRecursive(reflect.ValueOf(structPtr))
}

func floatRoundRecursive(val reflect.Value) {
	switch val.Kind() {
	case reflect.Ptr,
		reflect.Interface:
		value := val.Elem()
		if !value.IsValid() {
			return
		}
		floatRoundRecursive(value)
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			floatRoundRecursive(val.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			floatRoundRecursive(val.Index(i))
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			value := val.MapIndex(key)
			floatRoundRecursive(value)
		}
	case reflect.Float64:
		if val.CanSet() {
			val.SetFloat(numberutils.FloatRound(val.Float()))
		}
	}
}

// ConvertToBytes ...
func ConvertToBytes(buf *[]byte, obj interface{}) error {
	if buf == nil {
		return errors.New("nil buf convert")
	}

	typ := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if !val.IsValid() {
		return nil
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	kind := typ.Kind()
	switch {
	case typ.AssignableTo(timeTime):
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, &obj); err != nil {
			return err
		}
		*buf = append(*buf, buffer.Bytes()...)
	case kind == reflect.Map:
		keys := val.MapKeys()
		mapTyp := typ.Key().Kind()

		switch {
		case mapTyp == reflect.String:
			sort.SliceStable(keys, func(i, j int) bool {
				return keys[i].String() < keys[j].String()
			})
		case mapTyp >= reflect.Uint && mapTyp <= reflect.Uintptr:
			sort.SliceStable(keys, func(i, j int) bool {
				return keys[i].Uint() < keys[j].Uint()
			})
		case mapTyp >= reflect.Int && mapTyp <= reflect.Int64:
			sort.SliceStable(keys, func(i, j int) bool {
				return keys[i].Int() < keys[j].Int()
			})
		case mapTyp >= reflect.Uint && mapTyp <= reflect.Uintptr:
			sort.SliceStable(keys, func(i, j int) bool {
				return keys[i].Float() < keys[j].Float()
			})
		}

		for _, k := range keys {
			v := val.MapIndex(k)
			ConvertToBytes(buf, v.Interface())
		}

	case kind == reflect.Slice || kind == reflect.Array:
		vlen := val.Len()
		for i := 0; i < vlen; i++ {
			ConvertToBytes(buf, val.Index(i).Interface())
		}
	case kind == reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			ConvertToBytes(buf, val.Field(i).Interface())
		}
	default:
		buffer := new(bytes.Buffer)
		if err := rlp.Encode(buffer, &obj); err != nil {
			return err
		}
		*buf = append(*buf, buffer.Bytes()...)
	}

	return nil
}

func ParseHideTag(structPtr interface{}) {
	parseHideRecursive(reflect.ValueOf(structPtr))
}

func parseHideRecursive(val reflect.Value) {
	switch val.Kind() {
	case reflect.Ptr,
		reflect.Interface:
		value := val.Elem()
		if !value.IsValid() {
			return
		}
		parseHideRecursive(value)
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			tfield := val.Type().Field(i)
			vfield := val.Field(i)

			if tfield.Type.Kind() == reflect.String {
				tag := tfield.Tag.Get("hide")
				if len(tag) > 0 && vfield.CanSet() {
					str := vfield.String()
					vfield.SetString(fmt.Sprintf("%s%s%s", str[:2], strings.Repeat(tag, 4), str[len(str)-2:len(str)]))
				}
			} else {
				parseHideRecursive(val.Field(i))
			}
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			parseHideRecursive(val.Index(i))
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			value := val.MapIndex(key)
			parseHideRecursive(value)
		}
	}
}

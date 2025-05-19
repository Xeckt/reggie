package reggie

import (
	"fmt"
	"reflect"
	"strings"
)

func toBaseType(value any) (any, error) {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		return val.Convert(reflect.TypeOf("")).Interface().(string), nil
	case reflect.Slice:
		elemKind := val.Type().Elem().Kind()
		switch elemKind {
		case reflect.Uint8:
			return val.Convert(reflect.TypeOf([]byte(nil))).Interface().([]byte), nil
		case reflect.String:
			return val.Convert(reflect.TypeOf([]string(nil))).Interface().([]string), nil
		default:
			return nil, fmt.Errorf("unsupported slice element kind: %s", elemKind)
		}

	case reflect.Uint64:
		return val.Convert(reflect.TypeOf(uint64(0))).Interface().(uint64), nil

	case reflect.Uint32:
		return val.Convert(reflect.TypeOf(uint32(0))).Interface().(uint32), nil

	default:
		return nil, fmt.Errorf("unsupported kind: %s", val.Kind())
	}
}

func containsZeroByte(s string) bool {
	return strings.IndexByte(s, 0) != -1
}

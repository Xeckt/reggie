package reggie

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

func containsZeroByte(s string) bool {
	return strings.IndexByte(s, 0) != -1
}

// Obtains a value from the key `name`.
// It will get any type from the registry without needing
// to specify the specific registry.GetXValue(...) functions.
func (k *Key) GetValue(name string) (any, error) {
	var err error

	var v any

	_, t, _ := k.Handle.GetValue(name, nil)

	switch t {
	case registry.NONE:
		return nil, nil // Allow nil checks

	case registry.SZ:
		v, _, err = k.Handle.GetStringValue(name)

	case registry.EXPAND_SZ:
		v, _, err = k.Handle.GetStringValue(name)
		v, err = registry.ExpandString(v.(string))

	case registry.DWORD, registry.QWORD:
		v, _, err = k.Handle.GetIntegerValue(name)

	case registry.BINARY:
		v, _, err = k.Handle.GetBinaryValue(name)

	case registry.MULTI_SZ:
		v, _, err = k.Handle.GetStringsValue(name)
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

// Creates a value in accordance with the std registry package constraints.
// Underlying value type is reflected. Supports all known types of values.
func (k *Key) CreateValue(key string, value any) error {
	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		if containsZeroByte(value.(string)) {
			return fmt.Errorf("value for %q contains a zero byte, which is not allowed", key)
		}
		return k.Handle.SetStringValue(key, value.(string))

	case reflect.Slice:
		if reflect.TypeOf(value).Elem().Kind() == reflect.String {
			return k.Handle.SetStringsValue(key, value.([]string))
		} else if reflect.TypeOf(value).Elem().Kind() == reflect.Uint8 {
			return k.Handle.SetBinaryValue(key, value.([]byte))
		}

	case reflect.Uint64:
		return k.Handle.SetQWordValue(key, value.(uint64))

	case reflect.Uint32:
		return k.Handle.SetDWordValue(key, value.(uint32))

	default:
		return fmt.Errorf("Unsupported type %T", value)
	}

	return nil
}

// Safely checks if the value exists and deletes it.
func (k *Key) DeleteValue(name string) error {
	err := k.Handle.DeleteValue(name)
	if err != nil {
		if errors.Is(err, syscall.ENOENT) || strings.Contains(err.Error(), "The system cannot find the file specified") {
			return fmt.Errorf("value %q not found in key %s", name, k.Path)
		}
		return fmt.Errorf("failed to delete value %q in %s: %w", name, k.Path, err)
	}

	return nil
}

// Obtains all key=>value pairs from the specified key
func (k *Key) GetValueAndNames() (map[string]any, error) {
	valNames, err := k.Handle.ReadValueNames(-1)
	if err != nil {
		return nil, fmt.Errorf("Unable to read value names for %s: %w", k.Path, err)
	}

	values := make(map[string]any)

	for _, v := range valNames {
		valData, err := k.GetValue(v)
		if err != nil {
			return nil, fmt.Errorf("Unable to get value for %s: %w", v, err)
		}
		values[v] = valData
	}

	return values, nil
}

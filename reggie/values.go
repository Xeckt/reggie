package reggie

import (
	"errors"
	"fmt"
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
// Supports all known types of values.
func (k *Key) CreateValue(key string, value any, valueType uint32) error {
	var err error

	switch valueType {
	case registry.SZ:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("value is not of type string but of type: %T", value)
		}

		if containsZeroByte(v) {
			return fmt.Errorf("value for %q contains a zero byte, which is not allowed", key)
		}

		err = k.Handle.SetStringValue(key, v)
	case registry.EXPAND_SZ:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("value is not of type string but of type: %T", value)
		}

		if containsZeroByte(v) {
			return fmt.Errorf("value for %q contains a zero byte, which is not allowed", key)
		}

		err = k.Handle.SetExpandStringValue(key, v)
	case registry.MULTI_SZ:
		v, ok := value.([]string)
		if !ok {
			return fmt.Errorf("value is not of type string but of type: %T", value)
		}

		err = k.Handle.SetStringsValue(key, v)
	case registry.BINARY:
		v, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("value is not of type []byte but of type: %T", value)
		}

		err = k.Handle.SetBinaryValue(key, v)
	case registry.QWORD:
		v, ok := value.(uint64)
		if !ok {
			return fmt.Errorf("value is not of type uint64 but of type: %T", value)
		}

		err = k.Handle.SetQWordValue(key, v)
	case registry.DWORD:
		v, ok := value.(uint32)
		if !ok {
			return fmt.Errorf("value is not of type uint32 but of type: %T", value)
		}

		err = k.Handle.SetDWordValue(key, v)
	default:
		return fmt.Errorf("Unable to match case for CreateValue value %v", value)
	}

	if err != nil {
		return err
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

package reggie

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// Obtains a value from the key `name`.
// It will get any type from the registry without needing
// to specify the specific registry.GetXValue(...) functions.
func (k *Key) GetValue(name string) (any, error) {

	// Only need to obtain type here so do not provide a buffer
	_, t, _ := k.Handle.GetValue(name, nil)

	switch t {
	case registry.NONE:
		return nil, nil // Allow nil checks

	case registry.SZ:
		v, _, err := k.Handle.GetStringValue(name)
		return v, err

	case registry.EXPAND_SZ:
		v, _, err := k.Handle.GetStringValue(name)
		v, err = registry.ExpandString(v)
		return v, err

	case registry.DWORD, registry.QWORD:
		v, _, err := k.Handle.GetIntegerValue(name)
		return v, err

	case registry.BINARY:
		v, _, err := k.Handle.GetBinaryValue(name)
		return v, err

	case registry.MULTI_SZ:
		v, _, err := k.Handle.GetStringsValue(name)
		return v, err

	default:
		return nil, fmt.Errorf("Unable to get value from key name %s on open key %s", name, k.Path)
	}
}

// Creates a value in accordance with the std registry package constraints.
// Underlying value type is reflected. Supports all known types of values.
func (k *Key) CreateValue(key string, value any) error {

	t, err := toBaseType(value)
	if err != nil {
		return err
	}

	switch t.(type) {
	case string:
		if containsZeroByte(t.(string)) {
			return fmt.Errorf("value for %q contains a zero byte, which is not allowed", key)
		}
		return k.Handle.SetStringValue(key, t.(string))

	case []string:
		return k.Handle.SetStringsValue(key, t.([]string))

	case []byte:
		return k.Handle.SetBinaryValue(key, t.([]byte))

	case uint64:
		return k.Handle.SetQWordValue(key, t.(uint64))

	case uint32:
		return k.Handle.SetDWordValue(key, t.(uint32))

	default:
		return fmt.Errorf("Unsupported type %T", t)
	}
}

// Loops through the provided map and calls CreateValue(...) to create all key=>values
// in the current key object.
func (k *Key) CreateValueMany(data map[string]any) error {
	for key, value := range data {
		err := k.CreateValue(key, value)
		if err != nil {
			return fmt.Errorf("Error creating key: %s with value %v: %w", key, value, err)
		}
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

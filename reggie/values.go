package reggie

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func containsZeroByte(s string) bool {
	return strings.IndexByte(s, 0) != -1
}

// GetValue takes a specified registry key and returns the value of the named key `n`.
// This is a generic wrapper function over registry.GetValue
func (k *Key) GetValue(n string) (any, error) {
	var err error
	var v any
	_, t, _ := k.Handle.GetValue(n, nil)

	switch t {
	case registry.NONE:
		return nil, nil // Allow nil checks

	case registry.SZ:
		v, _, err = k.Handle.GetStringValue(n)

	case registry.EXPAND_SZ:
		v, _, err = k.Handle.GetStringValue(n)
		v, err = registry.ExpandString(v.(string))

	case registry.DWORD, registry.QWORD:
		v, _, err = k.Handle.GetIntegerValue(n)

	case registry.BINARY:
		v, _, err = k.Handle.GetBinaryValue(n)

	case registry.MULTI_SZ:
		v, _, err = k.Handle.GetStringsValue(n)
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

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
	}

	if err != nil {
		return err
	}

	return nil
}

func (k *Key) DeleteValue(value string) error {
	v, err := k.GetValue(value)
	if err != nil {
		return err
	}
	if v == nil {
		return fmt.Errorf("There is no value %s in %s", value, k.Path)
	}
	err = k.Handle.DeleteValue(value)
	if err != nil {
		return err
	}
	return nil
}

// GetValueAndNames() gets all key=>value pairs from the specified key
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

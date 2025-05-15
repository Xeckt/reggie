package reggie

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

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
		if _, ok := value.(string); !ok {
			return fmt.Errorf("value is not of type string but of type: %T", value)
		}
		err = k.Handle.SetStringValue(key, value.(string))
	case registry.EXPAND_SZ:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("value is not of type string but of type: %T", value)
		}
		err = k.Handle.SetExpandStringValue(key, value.(string))
	case registry.MULTI_SZ:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("value is not of type string but of type: %T", value)
		}
		err = k.Handle.SetStringsValue(key, value.([]string))
	case registry.BINARY:
		if _, ok := value.([]byte); !ok {
			return fmt.Errorf("value is not of type []byte but of type: %T", value)
		}
		err = k.Handle.SetBinaryValue(key, value.([]byte))
	case registry.QWORD:
		if _, ok := value.(uint64); !ok {
			return fmt.Errorf("value is not of type uint64 but of type: %T", value)
		}
		err = k.Handle.SetQWordValue(key, value.(uint64))
	case registry.DWORD:
		if _, ok := value.(uint32); !ok {
			return fmt.Errorf("value is not of type uint32 but of type: %T", value)
		}
		err = k.Handle.SetDWordValue(key, value.(uint32))
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
	if v != nil {
		err := k.Handle.DeleteValue(value)
		if err != nil {
			return err
		}
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

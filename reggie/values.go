package reggie

import "golang.org/x/sys/windows/registry"

// GetValue takes a specified registry key and returns the value of the named key `n`.
// This is a generic wrapper function over registry.GetValue
func GetValue(k registry.Key, n string) (any, error) {
	var err error
	var v any
	_, t, _ := k.GetValue(n, nil)

	switch t {
	case registry.NONE:
		return nil, nil // Allow nil checks

	case registry.SZ:
		v, _, err = k.GetStringValue(n)

	case registry.EXPAND_SZ:
		v, _, err = k.GetStringValue(n)
		v, err = registry.ExpandString(v.(string))

	case registry.DWORD, registry.QWORD:
		v, _, err = k.GetIntegerValue(n)

	case registry.BINARY:
		v, _, err = k.GetBinaryValue(n)

	case registry.MULTI_SZ:
		v, _, err = k.GetStringsValue(n)
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

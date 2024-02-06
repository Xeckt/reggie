//go:build windows

package reggie

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"strings"
)

type Reg struct {
	RootKey    registry.Key       // The key in which to access (HKLM, HKCU, etc). Can also be a subkey
	ActiveKey  registry.Key       // The key currently opened, if different from the root key (i.e parent key)
	Path       string             // The path inside RootKey to use
	Permission uint32             // The access type for given RootKey and Path
	SubKeyMap  map[string]*SubKey // Holds the subkeys underneath RootKey
}

type SubKey struct {
	Data  *Reg           // Holds the subkey information
	Value map[string]any // Holds the key value data stored in each subkey.
}

// NewReg initialises a Reg struct. Permission can be supplied via go's registry package.
func NewReg(permission uint32) *Reg {
	return &Reg{
		Permission: permission,
		SubKeyMap:  make(map[string]*SubKey),
	}
}

// NewSubKey initialises a struct for Reg.SubKeyMap.
func NewSubKey(permission uint32) *SubKey {
	return &SubKey{
		Data:  NewReg(permission),
		Value: make(map[string]any),
	}
}

// OpenKey is used to open a key inside the SubKey struct. Parameter `populateKeyValues` is true or false if you
// want to populate the SubKeyMap map with it's held data.
func (s *SubKey) OpenKey(populateKeyValues bool) (*Reg, error) {
	k := Reg{
		RootKey:    s.Data.RootKey,
		Path:       s.Data.Path,
		Permission: s.Data.Permission,
		SubKeyMap:  make(map[string]*SubKey),
	}
	if populateKeyValues {
		err := k.GetKeysValues()
		if err != nil {
			return nil, err
		}
	}
	return &k, nil
}

// GetKeysValues obtains RootKey, enumerates through each subkey in given Path. Each subkey will be attached inside Reg.SubKeyMap
// with its relevant data.
func (r *Reg) GetKeysValues() error {
	s, err := r.EnumerateSubKeys()
	if err != nil {
		return err
	}

	for _, subkey := range s {
		p := r.Path + "\\" + subkey

		key, err := registry.OpenKey(r.RootKey, p, r.Permission) // Must open each subkey as a new key
		if err != nil {
			if strings.Contains(err.Error(), "Access is denied") {
				return fmt.Errorf("access denied for key: %s", p)
			}
			return err
		}

		if r.SubKeyMap == nil {
			r.SubKeyMap = make(map[string]*SubKey)
		}

		if r.SubKeyMap[subkey] == nil {
			r.SubKeyMap[subkey] = NewSubKey(r.Permission)
		}

		r.SubKeyMap[subkey].Data = &Reg{
			Path:       p,
			RootKey:    r.RootKey,
			ActiveKey:  key,
			Permission: r.Permission,
		}

		names, err := key.ReadValueNames(0)
		if err != nil {
			return err
		}

		for _, n := range names {
			value, err := r.GetValue(key, n)
			if err != nil {
				return err
			}
			if value != nil { // Populate if it's not empty
				r.SubKeyMap[subkey].Value[n] = value
			}
		}

	}
	return nil
}

// GetValue takes a specified registry key and returns the value of the named key `n`
func (r *Reg) GetValue(k registry.Key, n string) (any, error) {
	var err error
	_, t, _ := k.GetValue(n, nil)
	var v any

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

// CreateKey creates a child key from the Reg.ActiveKey
func (r *Reg) CreateKey(name string) error {
	_, exists, err := registry.CreateKey(r.ActiveKey, name, r.Permission)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("key %s already exists", name)
	}
	err = r.GetKeysValues() // Repopulate struct with new keys
	if err != nil {
		return err
	}
	return nil
}

// CreateValue will create a designated key=value pair based on the valueType passed from registry type constants
func (r *Reg) CreateValue(key string, value any, valueType uint32) error {
	var err error

	switch valueType {
	case registry.SZ:
		err = r.ActiveKey.SetStringValue(key, value.(string))
	case registry.EXPAND_SZ:
		err = r.ActiveKey.SetExpandStringValue(key, value.(string))
	case registry.MULTI_SZ:
		err = r.ActiveKey.SetStringsValue(key, value.([]string))
	case registry.BINARY:
		err = r.ActiveKey.SetBinaryValue(key, value.([]byte))
	case registry.QWORD:
		err = r.ActiveKey.SetQWordValue(key, value.(uint64))
	case registry.DWORD:
		err = r.ActiveKey.SetDWordValue(key, value.(uint32))
	}

	if err != nil {
		return err
	}

	err = r.GetKeysValues()
	if err != nil {
		return err
	}

	return nil
}

// EnumerateSubKeys takes the given key in the Reg struct, enumerate
// and find it's subkeys. Amount specifies how many subkeys you want to enumerate.
// The default value is 0 to enumerate all within a given key, anything above 0 will enumerate to the specified amount.
// Amount cannot have more than one element. Behaves the same as specified in registry documentation: https://pkg.go.dev/golang.org/x/sys/windows/registry#Key.ReadSubKeyNames
func (r *Reg) EnumerateSubKeys(amount ...int) ([]string, error) {
	if len(amount) > 1 {
		return nil, errors.New("length of amount cannot exceed 1")
	}

	var sKeys []string
	key, err := registry.OpenKey(r.RootKey, r.Path, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, err
	}

	if len(amount) != 0 {
		sKeys, err = key.ReadSubKeyNames(amount[0])
	} else {
		sKeys, err = key.ReadSubKeyNames(0)
	}

	if err != nil {
		return nil, err
	}
	return sKeys, nil
}

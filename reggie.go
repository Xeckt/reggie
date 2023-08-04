package reggie

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
)

type Reg struct {
	RootKey     registry.Key       // The key in which to access (HKLM, HKCU, etc). Can also be a subkey
	Path        string             // The path inside RootKey to use
	Access      uint32             // The access type for given RootKey and Path
	CurrOpenKey registry.Key       // The key currently opened
	SubKeys     map[string]*SubKey // Holds the subkeys underneath RootKey
}

type SubKey struct {
	Key   *Reg           // Holds the subkey information
	Value map[string]any // Holds the key value data stored in each subkey.
}

func (s SubKey) GetKey() *Reg { return s.Key }

func (s SubKey) GetValue(name string) any { return s.Value[name] }

// New initialises a Reg struct with ALL_ACCESS permissions. Better used for testing unless requirements demand it.
func New() *Reg {
	return &Reg{
		Access:  registry.ALL_ACCESS,
		SubKeys: make(map[string]*SubKey),
	}
}

// GetSubKeysValues obtains RootKey, enumerates through each subkey in given Path, and obtains each non-empty value attached within every subkey.
// When successful, each subkey will be attached to *Reg.SubKeys and each subkeys key=value pair in *Reg.SubKeys[k].Value
func (r *Reg) GetSubKeysValues() error {
	s, err := r.EnumerateSubKeys()
	if err != nil {
		return err
	}
	for _, subkey := range s {
		p := r.Path + "\\" + subkey
		key, err := registry.OpenKey(r.RootKey, p, r.Access) // Must open each subkey as a new key
		if err != nil {
			fmt.Println(subkey, r.Path, err)
			return err
		}
		names, err := key.ReadValueNames(0)
		if err != nil {
			return err
		}
		for _, name := range names {
			if r.SubKeys[subkey] == nil {
				r.SubKeys[subkey] = &SubKey{
					Value: map[string]any{}, // Create a blank value map
				}
			}
			value, _ := r.GetValueFromType(key, name)
			if len(name) != 0 {
				r.SubKeys[subkey].Key.RootKey = key // Allow for an interactable subkey object inside each subkey map
				r.SubKeys[subkey].Key.Path = p
				r.SubKeys[subkey].Value[name] = value
			}
		}
	}
	return nil
}

// GetValueFromType takes a specified registry key and returns the value of the named key `n`
func (r *Reg) GetValueFromType(k registry.Key, n string) (any, error) {
	_, t, err := k.GetValue(n, nil)
	if err != nil {
		return nil, err
	}
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
	if v != nil {
		return v, nil
	}
	return v, err
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

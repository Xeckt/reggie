package main

import (
	"errors"
	"golang.org/x/sys/windows/registry"
)

type Reg struct {
	Key           registry.Key    // The key in which to access (HKLM, HKCU, etc)
	Path          string          // The path inside Key to use
	Access        int             // The access type for given Key and Path
	SubKeys       map[string]*Reg // Holds the subkeys underneath Key
	SubKeysValues map[string]any  // Holds the key value data stored in each subkey.
}

// New initialises a basic *Reg object to handle and begin using.
// Not recommended if not using for tests as Access is ALL_ACCESS, unless your
// requirements meet this demand.
func New() *Reg {
	return &Reg{
		Access:        registry.ALL_ACCESS,
		SubKeys:       make(map[string]*Reg),
		SubKeysValues: make(map[string]any),
	}
}

// GetSubKeysValues obtains Key, enumerates through each subkey in given Path, and obtains each key value attached within every subkey.
// When successful, each subkey will be attached to *Reg.SubKeys and each subkeys key=value pair in *Reg.SubKeysValues
func (r *Reg) GetSubKeysValues() error {
	s, _ := r.EnumerateSubKeys()
	for _, subkey := range s {
		p := r.Path + "\\" + subkey
		key, err := registry.OpenKey(r.Key, p, registry.ALL_ACCESS) // Must open each subkey as a new key
		if err != nil {
			return err
		}
		names, err := key.ReadValueNames(0)
		for _, n := range names {
			if len(n) != 0 {
				vt, _ := r.GetValueFromType(key, n)
				if err != nil {
					return err
				}
				r.SubKeysValues[n] = vt
			}
		}
		r.SubKeys[subkey] = &Reg{SubKeysValues: r.SubKeysValues}
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

// EnumerateSubKeys takes the given key in the Reg struct and will enumerate
// and find it's subkeys. Amount specifies how many subkeys you want to enumerate.
// The default value is 0 to enumerate all, anything above 0 will enumerate to the specified amount.
// behaves the same as specified in registry documentation: https://pkg.go.dev/golang.org/x/sys/windows/registry#Key.ReadSubKeyNames
// Amount cannot have more than one element.
func (r *Reg) EnumerateSubKeys(amount ...int) ([]string, error) {
	if len(amount) > 1 {
		return nil, errors.New("length of amount cannot exceed 1")
	}
	var sKeys []string
	key, err := registry.OpenKey(r.Key, r.Path, registry.ENUMERATE_SUB_KEYS)
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

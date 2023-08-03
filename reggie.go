package main

import (
	"errors"
	"golang.org/x/sys/windows/registry"
)

type Reg struct {
	Key           registry.Key    // The key in which to access (HKLM, HKCU, etc)
	Path          string          // The path inside Key to use
	Access        int             // The access type for given Key and Path
	subKeys       map[string]*Reg // Holds the subkeys underneath Key
	subKeysValues map[string]any  // Holds the key value data stored in each subkey.
}

// New initialises a basic *Reg object to handle and begin using.
// Not recommended if not using for tests as Access is ALL_ACCESS, unless your
// requirements meet this demand.
func New() *Reg {
	return &Reg{
		Access:        registry.ALL_ACCESS,
		subKeys:       make(map[string]*Reg),
		subKeysValues: make(map[string]any),
	}
}

// getSubKeysValues obtains Key, enumerates through each subkey in given Path, and obtains each key value attached within
// every subkey.
func (r *Reg) getSubKeysValues() error {
	s, _ := r.enumerateSubKeys()
	for _, subkey := range s {
		p := r.Path + "\\" + subkey
		key, err := registry.OpenKey(r.Key, p, registry.ALL_ACCESS)
		if err != nil {
			return err
		}
		names, err := key.ReadValueNames(0)
		for _, n := range names {
			if len(n) != 0 {
				vt, err := r.getValueFromType(n)
				if err != nil {
					return err
				}
				r.subKeysValues[n] = vt
			}
		}
		r.subKeys[subkey] = &Reg{subKeysValues: r.subKeysValues}
	}
	return nil
}

// getValueFromType takes a registry key defined in the Reg struct and a named data key by name, and will return
// the given value based on its registry type.
func (r *Reg) getValueFromType(n string) (any, error) {
	_, t, err := r.Key.GetValue(n, nil)
	if err != nil {
		return nil, err
	}
	var v any
	switch t {
	case registry.NONE:
		return nil, nil // Allow nil checks
	case registry.SZ:
		v, _, err = r.Key.GetStringValue(n)
	case registry.EXPAND_SZ:
		v, _, err = r.Key.GetStringValue(n)
		v, err = registry.ExpandString(v.(string))
	case registry.DWORD, registry.QWORD:
		v, _, err = r.Key.GetIntegerValue(n)
	case registry.BINARY:
		v, _, err = r.Key.GetBinaryValue(n)
	case registry.MULTI_SZ:
		v, _, err = r.Key.GetStringsValue(n)
	}
	if v != nil {
		return v, nil
	}
	return v, err
}

// enumerateSubKeys takes the given key in the Reg struct and will enumerate
// and find it's subkeys. Amount specifies how many subkeys you want to enumerate.
// The default value is 0 to enumerate all, anything above 0 will enumerate to the specified amount.
// behaves the same as specified in registry documentation: https://pkg.go.dev/golang.org/x/sys/windows/registry#Key.ReadSubKeyNames
// Amount cannot have more than one element.
func (r *Reg) enumerateSubKeys(amount ...int) ([]string, error) {
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

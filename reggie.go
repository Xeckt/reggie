package reggie

import (
	"errors"
	"golang.org/x/sys/windows/registry"
)

type Reg struct {
	RootKey    registry.Key // The key in which to access (HKLM, HKCU, etc). Can also be a subkey
	Path       string       // The path inside RootKey to use
	Permission uint32       // The access type for given RootKey and Path
	OpenedKey  registry.Key
	SubKeys    map[string]*SubKey // Holds the subkeys underneath RootKey
}

type SubKey struct {
	Key   *Reg           // Holds the subkey information
	Value map[string]any // Holds the key value data stored in each subkey.
}

// OpenKey is used to open a key inside the SubKey struct. Parameter `populateKeyValues` is true or false if you
// want to populate the SubKeys map with the subkeys data.
func (s SubKey) OpenKey(populateKeyValues bool) (*Reg, error) {
	k := Reg{
		RootKey:    s.Key.RootKey,
		Path:       s.Key.Path,
		Permission: s.Key.Permission,
		SubKeys:    map[string]*SubKey{},
	}
	if populateKeyValues {
		err := k.GetKeysValues()
		if err != nil {
			return nil, err
		}
	}
	return &k, nil
}

// New initialises a Reg struct with ALL_ACCESS permissions. Better used for testing unless requirements demand it.
func New() *Reg {
	return &Reg{
		Permission: registry.ALL_ACCESS,
		SubKeys:    make(map[string]*SubKey),
	}
}

// GetKeysValues obtains RootKey, enumerates through each subkey in given Path. Each subkey will be attached inside Reg.SubKeys
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
			return err
		}
		if r.SubKeys[subkey] == nil {
			r.SubKeys[subkey] = &SubKey{
				Key:   &Reg{},
				Value: map[string]any{},
			}
		}
		names, err := key.ReadValueNames(0)
		if err != nil {
			return err
		}
		for _, n := range names {
			value, err := r.GetValueFromType(key, n)
			if err != nil {
				return err
			}
			r.SubKeys[subkey].Value[n] = value
		}
		r.SubKeys[subkey].Key.Path = p
		r.SubKeys[subkey].Key.OpenedKey = key
		r.SubKeys[subkey].Key.RootKey = r.RootKey
		r.SubKeys[subkey].Key.Permission = r.Permission
	}
	return nil
}

// GetValueFromType takes a specified registry key and returns the value of the named key `n`
func (r *Reg) GetValueFromType(k registry.Key, n string) (any, error) {
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

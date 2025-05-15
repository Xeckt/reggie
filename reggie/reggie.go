package reggie

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type Key struct {
	Handle  registry.Key
	Path    string
	Subkeys map[string]*SubKey
	Values  map[string]any
	Loaded  bool
}

type SubKey struct {
	Name   string
	Values map[string]any
	Child  *Key
}

// Open opens an existing key.
func OpenKey(root registry.Key, path string, access uint32) (*Key, error) {
	handle, err := registry.OpenKey(root, path, access)
	if err != nil {
		return nil, fmt.Errorf("Unable to open key %s: %w", path, err)
	}
	return &Key{handle, path, nil, nil, false}, nil
}

// Create creates a new key. This function will error if the key already
// exists, unlike the std package version where it will silently open anyway.
// This is done to enforce correctness while acting as a guard
func (k *Key) CreateKey(path string, access uint32) (*Key, error) {
	handle, openedExisting, err := registry.CreateKey(k.Handle, path, access)
	if err != nil {
		return nil, fmt.Errorf("Unable to create key %s: %w", k.Path, err)
	}
	if openedExisting {
		return nil, fmt.Errorf("Unable to create key %s for %s: already exists", path, k.Path)
	}
	return &Key{handle, path, nil, nil, false}, nil
}

// LoadWithLimit will get the current key you have opened and then enumerate it
// for futher subkeys and it's children.
// `limit` is an integer to specify if you want to only load x amount of items.
func (k *Key) LoadWithLimit(limit int) error {
	if k.Loaded {
		return fmt.Errorf("Cannot load data for %s: already loaded", k.Path)
	}

	valData, err := k.GetValueAndNames() // Make sure we load the the current keys data and not just subkeys
	if err != nil {
		return err
	}

	if valData != nil {
		k.Values = valData
	}

	names, err := k.Handle.ReadSubKeyNames(-1)
	if err != nil {
		return err
	}

	numberLoaded := 0

	k.Subkeys = make(map[string]*SubKey)

	for _, name := range names {
		if limit > 0 && numberLoaded == limit {
			break
		}

		childPath := k.Path + `\` + name

		h, err := OpenKey(k.Handle, name, registry.READ)

		if err != nil {
			return err
		}

		child := h
		child.Path = childPath

		valData, err := child.GetValueAndNames()
		if err != nil {
			return err
		}

		k.Subkeys[name] = &SubKey{
			Name:   name,
			Child:  child,
			Values: valData,
		}

		k.Subkeys[name].Child.Loaded = true

		numberLoaded++
	}

	k.Loaded = true
	return nil
}

// Load calls LoadWithLimit(0) for an explicit way to load keys
// with no limit.
func (k *Key) Load() error {
	return k.LoadWithLimit(0)
}

// GetValueAndNames() gets all key=>value pairs from the specified key
func (k *Key) GetValueAndNames() (map[string]any, error) {
	valNames, err := k.Handle.ReadValueNames(-1)
	if err != nil {
		return nil, fmt.Errorf("Unable to read value names for %s: %w", k.Path, err)
	}

	values := make(map[string]any)

	for _, v := range valNames {
		raw, err := k.GetValue(v)
		if err != nil {
			return nil, fmt.Errorf("Unable to get value for %s: %w", v, err)
		}
		values[v] = raw
	}

	return values, nil
}

// Walk will recursively traverse all keys and subkeys
func (k *Key) Walk(fn func(k *Key) error) error {
	if err := fn(k); err != nil {
		return err
	}

	if !k.Loaded {
		if err := k.Load(); err != nil {
			return fmt.Errorf("walk failed to load subkeys: %w", err)
		}
	}

	for _, sub := range k.Subkeys {
		if sub.Child != nil {
			if err := sub.Child.Walk(fn); err != nil {
				return err
			}
		}
	}

	return nil
}

// Close closes the key
func (k *Key) Close() error {
	err := k.Handle.Close()
	if err != nil {
		return fmt.Errorf("Unable to close key %s: %w", k.Path, err)
	}
	return nil
}

package reggie

import (
	"fmt"
	"maps"

	"golang.org/x/sys/windows/registry"
)

// Represents a registry key layout
type Key struct {
	Handle     registry.Key       // Parent registry key
	Path       string             // Path of the parent registry key
	Subkeys    map[string]*SubKey // Map of subkeys from the opened registry key
	Values     map[string]any     // Values inside the parent registry key
	Permission uint32             // Permission used for the key
	Loaded     bool               // Represents if the key has had its values / subkeys loaded
}

// Represents a registry subkey layout
type SubKey struct {
	Name   string         // Name of the subkey
	Values map[string]any // Values inside the subkey
	Child  *Key           // Child keys inside the subkey
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
		k.Values = make(map[string]any, len(valData))
		maps.Copy(k.Values, valData)
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

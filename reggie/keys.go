package reggie

import (
	"fmt"
	"maps"

	"golang.org/x/sys/windows/registry"
)

// Represents a registry key layout
type Key struct {
	Handle     registry.Key       // Parent registry key
	RootKey    registry.Key       // The top-most level key, e.g. HKCU, HKLM
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

// Opens an existing registry key.
func OpenKey(root registry.Key, path string, access uint32) (*Key, error) {
	handle, err := registry.OpenKey(root, path, access)
	if err != nil {
		return nil, fmt.Errorf("Unable to open key %s: %w", path, err)
	}
	return &Key{handle, root, path, nil, nil, access, false}, nil
}

// Creates a new key. This function will error if the key already
// exists, unlike the std package version where it will silently open regardless.
// Enforces correctness and guards.
func (k *Key) CreateKey(path string, access uint32) (*Key, error) {
	handle, openedExisting, err := registry.CreateKey(k.Handle, path, access)
	if err != nil {
		return nil, fmt.Errorf("Unable to create key %s: %w", k.Path, err)
	}
	if openedExisting {
		return nil, fmt.Errorf("Unable to create key %s for %s: already exists", path, k.Path)
	}
	return &Key{handle, k.RootKey, path, nil, nil, access, false}, nil
}

// Performs a deep in memory copy of current key
// and returns it
func (k *Key) CloneKey() (*Key, error) {
	clone := &Key{
		Path:       k.Path,
		Loaded:     true,
		Permission: k.Permission,
		Subkeys:    make(map[string]*SubKey, len(k.Subkeys)),
		Values:     make(map[string]any, len(k.Values)),
	}

	clone.Handle = k.Handle

	maps.Copy(clone.Values, k.Values)

	for k, v := range k.Subkeys {
		subClone := &SubKey{
			Name:   v.Name,
			Values: make(map[string]any, len(v.Values)),
		}

		maps.Copy(subClone.Values, v.Values)

		if v.Child != nil {
			c, err := v.Child.CloneKey()
			if err != nil {
				return nil, err
			}

			subClone.Child = c
		}

		clone.Subkeys[k] = subClone
	}
	return clone, nil
}

// Checks if the key exists and deletes it.
// Will return an error if the key does not exist.
// Do not use this on keys that have subkeys. For this, use DeleteKeysAll()
func (k *Key) DeleteKey(path string) error {
	if !KeyExists(k.Handle, path) {
		return fmt.Errorf("Cannot delete key %s, does not exist", path)
	}
	err := registry.DeleteKey(k.Handle, path)
	if err != nil {
		return fmt.Errorf("Error deleting key %s/%s: %w", k.Path, path, err)
	}
	return nil
}

// KeyExists checks if a registry key exists at the given path and root.
// It returns true if the key exists and can be opened, false otherwise.
func KeyExists(key registry.Key, path string) bool {
	k, err := registry.OpenKey(key, path, registry.READ)
	if err != nil {
		return false
	}
	_ = k.Close()
	return true
}

// Close closes the key
func (k *Key) Close() error {
	err := k.Handle.Close()
	if err != nil {
		return fmt.Errorf("Unable to close key %s: %w", k.Path, err)
	}
	return nil
}

package reggie

import (
	"fmt"
	"maps"

	"golang.org/x/sys/windows/registry"
)

// Open opens an existing registry key.
func OpenKey(root registry.Key, path string, access uint32) (*Key, error) {
	handle, err := registry.OpenKey(root, path, access)
	if err != nil {
		return nil, fmt.Errorf("Unable to open key %s: %w", path, err)
	}
	return &Key{handle, path, nil, nil, access, false}, nil
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
	return &Key{handle, path, nil, nil, access, false}, nil
}

// CloneKey performs a deep in memory copy of current key
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

// Close closes the key
func (k *Key) Close() error {
	err := k.Handle.Close()
	if err != nil {
		return fmt.Errorf("Unable to close key %s: %w", k.Path, err)
	}
	return nil
}

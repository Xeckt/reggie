package reggie

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type Key struct {
	Handle  registry.Key
	Path    string
	Subkeys map[string]*SubKey
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
	return &Key{handle, path, nil, false}, nil
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
	return &Key{handle, path, nil, false}, nil
}

// Load will get the current key you have opened and then enumerate through it,
// Finding further subkeys and values.
func (k *Key) Load(limit int) error {
	if k.Loaded {
		return fmt.Errorf("Cannot load data for %s: already loaded", k.Path)
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
			continue
		}

		child := &Key{
			Handle: h.Handle,
			Path:   childPath,
		}

		values := make(map[string]any)
		valNames, _ := h.Handle.ReadValueNames(-1)

		for _, v := range valNames {
			raw, _ := GetValue(h.Handle, v)
			values[v] = raw
		}

		k.Subkeys[name] = &SubKey{
			Name:   name,
			Child:  child,
			Values: values,
		}

		numberLoaded++
	}

	k.Loaded = true
	return nil
}

func (k *Key) Walk(fn func(k *Key) error) error {
	if err := fn(k); err != nil {
		return err
	}

	if !k.Loaded {
		if err := k.Load(0); err != nil {
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

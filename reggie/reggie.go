package reggie

import (
	"fmt"
	"maps"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Will get the current key you have opened and then enumerate it
// for futher subkeys and it's children.
// `limit` is an integer to specify if you want to only load x amount of items,
// not to be confused with loading x amount of keys inside keys
// from the current key. LoadWithLimit will only do a shallow read one level down.
// If you want full, recursive deep loading see DeepLoad
func (k *Key) LoadWithLimit(limit int) error {
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

// Recursively loads the current key and all of its descendant subkeys.
// It calls Load() on the current key if it has not been loaded yet, then traverses all
// loaded subkeys and attempts to load each child key recursively.
func (k *Key) DeepLoad() error {
	if err := k.Load(); err != nil {
		return err
	}

	for _, sub := range k.Subkeys {
		if sub.Child != nil {
			if err := sub.Child.DeepLoad(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Calls LoadWithLimit(0) for an explicit way to load keys
// with no limit.
func (k *Key) Load() error {
	return k.LoadWithLimit(0)
}

// Recursively traverses the key and all of its loaded subkeys from the top down,
// applying the provided function `fn` to each *Key in depth-first order.
//
// If the key has not been loaded yet, Walk will call Load() automatically
// to populate subkeys before continuing traversal.
//
// If fn returns an error at any point, Walk stops immediately and returns
// that error.
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

// Recursively traverses the key and all of its loaded subkeys in bottom up
// post order. Uses the same logic as Walk()
func (k *Key) WalkReverse(fn func(k *Key) error) error {
	if !k.Loaded {
		if err := k.Load(); err != nil {
			return fmt.Errorf("walk failed to load subkeys: %w", err)
		}
	}

	// Traverse subkeys first (post-order)
	for _, sub := range k.Subkeys {
		if sub.Child != nil {
			if err := sub.Child.WalkReverse(fn); err != nil {
				return err
			}
		}
	}

	// Then apply the function to the current key
	if err := fn(k); err != nil {
		return err
	}

	return nil
}

// Returns parent path and segment
// e.g., Software\MyApp\Config -> (Software\MyApp, Config)
func splitParent(path string) (string, string) {
	idx := strings.LastIndex(path, `\`)
	if idx == -1 {
		return "", path
	}
	return path[:idx], path[idx+1:]
}

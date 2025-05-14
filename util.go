//go:build windows

package reggie

import (
	"fmt"
)

// Traverse allows you to recursively traverse a given reg object based on its current key.
func Traverse(r *Reg, getKeyValues bool, fn func(reg *Reg)) error {
	if fn == nil {
		return fmt.Errorf("function is nil, traversal cannot continue")
	}

	for _, s := range r.SubKeyMap {
		r, err := s.OpenKey(getKeyValues)
		if err != nil {
			return err
		}

		fn(r)
		_ = Traverse(r, getKeyValues, fn)
	}

	return nil
}

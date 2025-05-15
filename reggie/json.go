package reggie

import "encoding/json"

type keyJson struct {
	Path       string                 `json:"path"`
	Values     map[string]any         `json:"values,omitempty"`
	Subkeys    map[string]*subKeyJson `json:"subkeys,omitempty"`
	Permission uint32                 `json:"permission"`
	Loaded     bool                   `json:"loaded"`
}

type subKeyJson struct {
	Name   string         `json:"name"`
	Values map[string]any `json:"values,omitempty"`
	Child  *keyJson       `json:"child,omitempty"`
}

// Marshals the current key object into JSON.
func (k *Key) ExportJson() ([]byte, error) {
	return json.MarshalIndent(convertKey(k), "", " ")
}

func convertKey(k *Key) *keyJson {
	export := &keyJson{
		Path:   k.Path,
		Loaded: k.Loaded,
		Values: k.Values,
	}

	if k.Subkeys != nil && len(k.Subkeys) > 0 {
		export.Subkeys = make(map[string]*subKeyJson)
		for k, v := range k.Subkeys {
			data := &subKeyJson{
				Name:   v.Name,
				Values: v.Values,
			}
			if v.Child != nil {
				data.Child = convertKey(v.Child)
			}
			export.Subkeys[k] = data
		}
	}

	return export
}

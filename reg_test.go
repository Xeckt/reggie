package reggie

import (
	"golang.org/x/sys/windows/registry"
	"strings"
	"testing"
)

var (
	r = New()
)

func TestReg_EnumerateSubKeys(t *testing.T) {
	r.RootKey = registry.CURRENT_CONFIG
	r.Path = `System`
	want := "CurrentControlSet"
	got, err := r.EnumerateSubKeys(1)
	if err != nil {
		t.Error(err)
	}
	if !strings.EqualFold(want, got[0]) {
		t.Error("Expected:", want, "Received:", got)
	}
}

func TestReg_GetValueFromType(t *testing.T) {
	r.RootKey = registry.LOCAL_MACHINE
	r.Path = `SYSTEM\CurrentControlSet\Control`
	subKey := `CurrentUser`
	key, err := registry.OpenKey(r.RootKey, r.Path, registry.ALL_ACCESS)
	if err != nil {
		t.Error("Could not open subkey", r.Path, "Error:", err)
	}
	got, err := r.GetValueFromType(key, subKey)
	if err != nil {
		t.Error(err)
	}
	want := "USERNAME"
	if want != got {
		t.Error("Expected:", want, "Received:", got)
	}
}

func TestReg_GetSubKeysValues(t *testing.T) {
	r.RootKey = registry.LOCAL_MACHINE
	r.Path = `SYSTEM\CurrentControlSet`
	err := r.GetKeysValues()
	if err != nil {
		t.Error(err)
	}
	want := []string{"Control", "CurrentUser", "USERNAME"}
	received := []string{}

	for subkey, keys := range r.SubKeys {
		for key, value := range keys.Value {
			if strings.EqualFold(want[0], subkey) &&
				strings.EqualFold(want[1], key) &&
				strings.EqualFold(want[2], value.(string)) {
				received = append(received, subkey, key, value.(string))
				break
			}
		}
	}
	for i, v := range want {
		if v != received[i] {
			t.Error("Received:", received[i], "Wanted:", v)
		}
	}
}

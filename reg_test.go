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
	r.Key = registry.CURRENT_CONFIG
	r.SubKeyPath = `System`
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
	r.Key = registry.LOCAL_MACHINE
	r.SubKeyPath = `SYSTEM\CurrentControlSet\Control`
	keyName := `CurrentUser`
	err := r.OpenKey()
	if err != nil {
		t.Error("Could not open subkey", r.SubKeyPath, "Error:", err)
	}
	got, err := r.GetValueFromType(r.CurrOpenKey, keyName)
	if err != nil {
		t.Error(err)
	}
	want := "USERNAME"
	if want != got {
		t.Error("Expected:", want, "Received:", got)
	}
}

func TestReg_GetSubKeysValues(t *testing.T) {
	r.Key = registry.LOCAL_MACHINE
	r.SubKeyPath = `SYSTEM\CurrentControlSet`
	err := r.OpenKey()
	if err != nil {
		t.Error("Could not open given key. Error:", err)
	}
	err = r.GetSubKeysValues()
	if err != nil {
		t.Error(err)
	}
	want := []string{"Control", "CurrentUser", "USERNAME"}
	received := []string{}
	for subkey, vMap := range r.SubKeys {
		for key, got := range vMap.SubKeysValues {
			if strings.EqualFold(subkey, "Control") &&
				strings.EqualFold(key, "CurrentUser") &&
				strings.EqualFold("USERNAME", got.(string)) {
				received = append(received, subkey, key, got.(string))
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

package reggie_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/Xeckt/reggie"
	"golang.org/x/sys/windows/registry"
)

const testPath = `ReggieTest`

func setupTestKey(t *testing.T) *reggie.Key {
	key, err := reggie.OpenKey(registry.CURRENT_USER, `Software`, registry.ALL_ACCESS)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if reggie.KeyExists(key.Handle, testPath) {
		fmt.Println("should?")
		key.DeleteKey(testPath)
	}

	testKey, err := key.CreateKey("ReggieTest", key.Permission)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := key.DeleteKey(testPath)
		if err != nil {
			log.Fatal(err)
		}
		key.Close()
		testKey.Close()
	})

	return testKey
}

func TestSetAndGetValue(t *testing.T) {
	key := setupTestKey(t)

	// Set
	err := key.CreateValue("TestString", "Hello World")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get
	val, err := key.GetValue("TestString")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if val.(string) != "Hello World" {
		t.Errorf("expected 'Hello World', got %v", val)
	}
}

func TestDeleteValue(t *testing.T) {
	key := setupTestKey(t)

	err := key.CreateValue("ToDelete", "DeleteMe")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	err = key.DeleteValue("ToDelete")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	v, err := key.GetValue("ToDelete")
	if v != nil {
		t.Error("Expected value to be nil, got: %w", v)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestLoad(t *testing.T) {
	key := setupTestKey(t)
	_, err := key.CreateKey("Child", registry.ALL_ACCESS)
	if err != nil {
		t.Fatalf("create subkey failed: %v", err)
	}

	err = key.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if key.Subkeys["Child"] == nil {
		t.Error("expected subkey 'Child' to be loaded")
	}

	err = key.DeleteKey("Child")
	if err != nil {
		t.Error(err)
	}
}

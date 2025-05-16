package examples

import (
	"fmt"
	"log"

	"github.com/Xeckt/reggie"
	"golang.org/x/sys/windows/registry"
)

func walkKeys() {
	key, err := reggie.OpenKey(registry.CURRENT_USER, `Software`, registry.READ)
	if err != nil {
		log.Fatalf("Failed to open root key: %v", err)
	}
	defer key.Close()

	key.DeepLoad()

	err = key.Walk(func(k *reggie.Key) error {
		fmt.Printf("Path: %s\n", k.Path)
		for name, value := range k.Values {
			fmt.Printf("  %s: %v\n", name, value)
		}
		return nil // return an error to halt traversal
	})

	if err != nil {
		log.Fatalf("Walk failed: %v", err)
	}
}

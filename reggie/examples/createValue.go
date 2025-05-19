package examples

import (
	"log"

	"github.com/Xeckt/reggie"
	"golang.org/x/sys/windows/registry"
)

func main() {
	key, err := reggie.OpenKey(registry.CURRENT_USER, `Software\ReggieDemo`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}
	defer key.Close()

	err = key.CreateValue("string key", uint64(12345))
	if err != nil {
		log.Fatal(err)
	}
}

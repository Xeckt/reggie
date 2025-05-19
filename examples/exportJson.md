```go
package examples

import (
	"fmt"
	"log"

	"github.com/Xeckt/reggie"
	"golang.org/x/sys/windows/registry"
)

func exportJson() {
	key, err := reggie.OpenKey(registry.CURRENT_USER, `Control Panel\Accessibility`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}
	defer key.Close()

	err = key.DeepLoad()
	if err != nil {
		log.Fatal(err)
	}

	json, err := key.ExportJson()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(json))
}
```
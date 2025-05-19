package examples

import (
	"log"

	"github.com/Xeckt/reggie"
	"golang.org/x/sys/windows/registry"
)

type QWORD uint64

func createValueMany() {
	reggieDemo, err := reggie.OpenKey(registry.CURRENT_USER, `Software\ReggieDemo`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}
	defer reggieDemo.Close()

	many := make(map[string]any)

	var v QWORD
	v = 239823982

	many["numberone"] = uint32(20)
	many["numbertwo"] = uint32(30)
	many["numberthree"] = "hellothere"
	many["numberfour"] = []byte("bigman")
	many["numberfive"] = v

	err = reggieDemo.CreateValueMany(many)
	if err != nil {
		log.Fatal(err)
	}
}

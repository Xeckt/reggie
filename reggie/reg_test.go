package reggie

import (
	"golang.org/x/sys/windows/registry"
)

var (
	r = NewReg(registry.ALL_ACCESS)
)

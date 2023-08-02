package main

import (
	"golang.org/x/sys/windows/registry"
	"log"
)

type Reg struct {
	Key    registry.Key
	Path   string
	Access int
	sKey   map[string]*Reg
	kV     map[string]any
}

func New() *Reg {
	return &Reg{
		sKey: make(map[string]*Reg),
		kV:   make(map[string]any),
	}
}

func (r *Reg) GetSubsKeysAndValues() {
	s := r.EnumerateSubKeysNames()
	for _, subkey := range s {
		p := r.Path + "\\" + subkey
		key, err := registry.OpenKey(r.Key, p, registry.ALL_ACCESS)
		if err != nil {
			log.Fatal(err)
		}
		names, err := key.ReadValueNames(0)
		for _, n := range names {
			if len(n) != 0 {
				vt := r.GetValueFromType(key, n)
				r.kV[n] = vt
			}
		}
		r.sKey[subkey] = &Reg{kV: r.kV}
	}
}

func (r *Reg) GetValueFromType(k registry.Key, n string) any {
	_, t, err := k.GetValue(n, nil)
	if err != nil {
		log.Fatal(err)
	}
	var v any
	switch t {
	case registry.NONE:
	case registry.SZ:
		v, _, err = k.GetStringValue(n)
		if err != nil {
			log.Fatal(err)
		}
	case registry.EXPAND_SZ:
		v, _, err = k.GetStringValue(n)
		v, err = registry.ExpandString(v.(string))
		if err != nil {
			log.Fatal(err)
		}
	case registry.DWORD, registry.QWORD:
		v, _, err = k.GetIntegerValue(n)
		if err != nil {
			log.Fatal(err)
		}
	case registry.BINARY:
		v, _, err = k.GetBinaryValue(n)
		if err != nil {
			log.Fatal(err)
		}
	case registry.MULTI_SZ:
		v, _, err = k.GetStringsValue(n)
		if err != nil {
			log.Fatal(err)
		}
	}
	return v
}

func (r *Reg) EnumerateSubKeysNames(level ...int) []string {
	if len(level) > 1 {
		log.Fatal("Level cannot exceed length of 1")
	}
	var sKeys []string
	key, err := registry.OpenKey(r.Key, r.Path, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		log.Fatal(err)
	}
	if len(level) != 0 {
		sKeys, err = key.ReadSubKeyNames(level[0])
	} else {
		sKeys, err = key.ReadSubKeyNames(0)
	}
	if err != nil {
		log.Fatal(err)
	}
	return sKeys
}

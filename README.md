# reggie
[![Go Reference](https://pkg.go.dev/badge/pkg.go.dev/github.com/Xeckt/reggie.svg)](https://pkg.go.dev/github.com/Xeckt/reggie)

A small wrapper over Go's std [registry](https://pkg.go.dev/golang.org/x/sys/windows/registry) package.

Documentation and examples are available here: [Documentation](#Documentation)
# Summary
[Go's registry package](https://pkg.go.dev/golang.org/x/sys/windows/registry) is extremely useful but limitations arise where
you have a wider range of requirements, thus requiring customised functions to handle the use case. 

Reggie assists on that front and gives you wider, structured utilities and access for handling the registry.

# Using Reggie
```
go get github.com/Xeckt/reggie
```

# Examples
Let's take an example where you want to access a *set* of registry keys, in the std package, it would look like:
```go
func main() {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Control Panel`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}
	subkeys, err := key.ReadSubKeyNames(0)
	for _, subkey := range subkeys {
		key, _, err := registry.CreateKey(registry.CURRENT_USER, `Control Panel\`+subkey, registry.ALL_ACCESS)
		if err != nil {
			log.Fatal(err)
		}
		moreKeys, err := key.ReadSubKeyNames(0)
		if err != nil {
			log.Fatal(err)
		}
		for _, k := range moreKeys {
			fmt.Println(subkey, "->", k)
		}
	}
}
```
It's quite something, additionally you would need to add your own magic. We will use `reggie` to do the exact same as the above:
```go
func main() {
	r := reggie.NewReg(registry.ALL_ACCESS)
	r.RootKey = registry.CURRENT_USER
	r.Path = `Control Panel`
	err := r.GetKeysValues() // Reggie will populate its own structs
	if err != nil {
		log.Fatal(err)
	}
	err = r.SubKeyMap["Accessibility"].Data.GetKeysValues()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Control panel subkeys -> ", r.SubKeyMap, "\nControl Panel - Accessibility Subkeys ->",
		r.SubKeyMap["Accessibility"].Data)
}
```
If we wanted to shorten that further, `reggie` has a recursive function `Traverse()`, which allows you to do bulk actions:
```go
func main() {
	r := reggie.NewReg(registry.ALL_ACCESS)
	r.RootKey = registry.CURRENT_USER
	r.Path = `Control Panel`
	err := r.GetKeysValues() // Reggie will populate its own structs
	if err != nil {
		log.Fatal(err)
	}
	reggie.Traverse(r, true, func(reg *reggie.Reg) {
		for key, value := range reg.SubKeyMap {
            fmt.Println(reg.Path, "->", key, value)
		}
	})
}
```
Given the registry path above, output would look like:
```
Control Panel\Accessibility -> ToggleKeys &{0xc0003fd620 map[Flags:62]}
Control Panel\Accessibility -> HighContrast &{0xc0003fce40 map[Flags:126 High Contrast Scheme: Previous High Contrast Scheme MUI Value:]}
Control Panel\Accessibility -> On &{0xc0003fd1a0 map[Locale:0 On:0]}
Control Panel\Accessibility -> SlateLaunch &{0xc0003fd320 map[ATapp:narrator LaunchAT:1]}
Control Panel\Accessibility -> Blind Access &{0xc0003fcd80 map[On:0]}
Control Panel\Accessibility -> ShowSounds &{0xc0003fd260 map[On:0]}
Control Panel\Accessibility -> SoundSentry &{0xc0003fd3e0 map[FSTextEffect:0 Flags:2 TextEffect:0 WindowsEffect:1]}
Control Panel\Accessibility -> Keyboard Preference &{0xc0003fcf30 map[On:0]}
...
```
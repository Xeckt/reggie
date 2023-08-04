# reggie
A small wrapper over Go's std [registry](https://pkg.go.dev/golang.org/x/sys/windows/registry) package.

[![Go Reference](https://pkg.go.dev/badge/pkg.go.dev/github.com/Xeckt/reggie.svg)](https://pkg.go.dev/github.com/Xeckt/reggie)

# Summary
Through working on various personal projects heavily reliant to Windows, some of the standard functions of Go's [registry](https://pkg.go.dev/golang.org/x/sys/windows/registry)
did not extend or perform the way I needed them to. Being low level, I was missing some handy behaviour I had
to end up writing myself.
There is a lot of manual behaviour necessary to write if you're only using the std registry package for your project which can become
overbearing and difficult to track, whereas
reggie will help ease the process of access and usage. It will (hopefully) help track the confusion of registry in your codebase.

The aim is to make it small, simple, and applicable for code readability.

# How to use
First get the package
```
go get github.com/Xeckt/reggie
```

Usage:
```go
// Grabs values of subkey located in specified Key
r := reggie.New()
r.Key = registry.LOCAL_MACHINE
r.Path = `System\CurrentControlSet`
err := r.GetSubKeysValues()
if err != nil {
	log.Fatal(err)
}
for key, subkey := range r.SubKeys {
	fmt.Println(key, subkey.Value)
}
```
The above usage example is basic. There are many ways to utilise reggie, one of them being subkey interaction
like a normal `registry.Key`. Let's take the subkey `TeamViewer`, which we first enumerate from `HKLM\LOCAL_MACHINE` on subkey path `SOFTWARE`
```go
r := reggie.New()
r.RootKey = registry.LOCAL_MACHINE
r.Path = `SOFTWARE`
err := r.GetKeysValues()
if err != nil {
	log.Fatal(err)
}
teamviewerKey, err := r.SubKeys["TeamViewer"].OpenKey()
if err != nil {
	log.Fatal(err)
}
fmt.Println(teamviewerKey.Value["Version"])
fmt.Println(teamviewerKey.Key.Path)
```


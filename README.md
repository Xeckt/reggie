# reggie
A small wrapper over Go's std [registry](https://pkg.go.dev/golang.org/x/sys/windows/registry) package.

Documentation and examples are available here: [Documentation](#Documentation)
# Summary
[Go's registry package](https://pkg.go.dev/golang.org/x/sys/windows/registry) becomes a bit tedious when handling
a lot of keys inside the registry at once, and fetching data from them. 

So, reggie helps handle situations where you are dealing with larger sets of registry data.

# Using Reggie
```
go get github.com/Xeckt/reggie
```

# Documentation
[![Go Reference](https://pkg.go.dev/badge/pkg.go.dev/github.com/Xeckt/reggie.svg)](https://pkg.go.dev/github.com/Xeckt/reggie)

You usually start by defining a `Reggie.Reg` struct object, either with the `New()` func or by populating the struct yourself.
It uses the same format as the [registry](https://pkg.go.dev/golang.org/x/sys/windows/registry) package with minor additions.
```go
r := reggie.New()
r.Key = registry.LOCAL_MACHINE
r.Path = `System\CurrentControlSet`
```
From here you have few options. Reggie can enumerate through the subkey names for you such as the provided location above, but we will take advantage
of the `GetSubKeysValues()` function. Which will populate the `SubKeys` map in our struct with the specified registry locations
key data.
```go
err := r.GetSubKeysValues()
if err != nil {
	log.Fatal(err)
}
for key, subkey := range r.SubKeys {
    fmt.Println(key, subkey.Value)
}
```
If you know what you're looking for specifically, you have a few approaches, and reggie still allows you to utilise the [std registry package functions](https://pkg.go.dev/golang.org/x/sys/windows/registry):
```go
r := reggie.New()
r.RootKey = registry.LOCAL_MACHINE
r.Path = `SOFTWARE`
err := r.GetKeysValues()
if err != nil {
	log.Fatal(err)
}
teamviewer := r.SubKeys["TeamViewer"]
fmt.Println(teamviewer.Value["Version"])
fmt.Println(teamviewer.Key.OpenedKey.GetStringValue("Version"))
```
We could take this a step further by trailing into `TeamViewer` subkeys as well, by taking advantage of the `OpenKey(bool)` function
and setting its parameter to true, otherwise it creates an empty subkey map for you to handle while keeping related key objects. Useful when you are
handling situations where string magic is necessary:
```go
s, err := r.SubKeys["TeamViewer"].OpenKey(true)
if err != nil {
	log.Fatal(err)
}
fmt.Println(s.SubKeys) // Returns a map of the subkeys you can use as normal registry.Key objects
```
A prominent feature of reggie is the `Traverse()` function. Sometimes you don't know what you're looking for, or you may have dynamic
data you need to handle. This function will recursively traverse the tree below a given `Key`. It helps you write your own magic:
```go
err = reggie.Traverse(r, false, func(reg *reggie.Reg) {
	if strings.Contains(reg.Path, "QNAP") {
		fmt.Println(reg.Path)
		return
	}
})
```
Something that reggie doesn't provide yet, is the handling of the errors that are under the hood in the `registry` package. Because the traversal can start
from the bottom up or the up down, you may experience different results each time it's run. So running the same code above may result in error one time and not another. 
This is also rather complicated considering the errors are inherently returned by Windows, so a lot of string magic needs to happen.

A solution to this is being worked on currently.
# reggie
A small wrapper over Go's std [Registry](https://pkg.go.dev/golang.org/x/sys/windows/registry) package.

[![Go Reference](https://pkg.go.dev/badge/pkg.go.dev/github.com/Xeckt/reggie.svg)](pkg.go.dev/github.com/Xeckt/reggie)

# Why?
Through working on various personal projects heavily related to Windows, some of the standard functions
did not extend or perform the way I needed them to. Being low level, I was missing some handy behaviour I had
to end up writing myself.

The aim is not to over-complicate the usage, but make it small, simple, and apply more readability to the code.

It can only search through one level of `subkeys` from a given `key`, however, I aim to add the behaviour of crawling
multiple levels of registry subkeys.

Still in development, don't expect it to be perfect.
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



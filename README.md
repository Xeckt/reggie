# reggie
<img src="./logo.png" width=250 height =250>

A clean, zero dependency wrapper around Go's `golang.org/x/sys/windows/registry` package.  

---

## Features

- [x] Effortlessly open, create, edit, or delete keys with guards in place

- [x] JSON exports at the object level. Export any registry key object as structured, readable JSON — perfect for API's, dumping, etc.

- [x] Seamlessly walk through registry hierarchies using a depth-first pattern.

- [x] Dynamically create values with automatic type enforcement and correct constraints.

- [x] Easily fetch all key-value pairs from any key in one call.

- [x] Retrieve typed values easily without boilerplate.

- [x] Clone a key object in memory, (including all values and subkeys) without affecting the original — useful for testing or state comparisons.

---

## Examples

See [examples folder](./reggie/examples/)

Since `v1.0.0` there are more features you can utilise. 

For example, if you want to expose registry information over an API, you can export the data sets to JSON:

```go
func main() {
	key, err := reggie.OpenKey(registry.CURRENT_USER, `Control Panel\Accessibility`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}
	defer key.Close()

	_ = key.DeepLoad() // There are 3 different load functions. Check the docs.

	json, err := key.ExportJson() // Built in functions to convert the data structure into a JSON compatible format
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(json))
}
```

Maybe you want to apply your own logic while walking through the registry:

```go
func main() {
	key, err := reggie.OpenKey(registry.CURRENT_USER, `Software`, registry.READ)
	if err != nil {
		log.Fatalf("Failed to open root key: %v", err)
	}
	defer key.Close()

	key.DeepLoad() // DeepLoad first if you want an easy traversal. Optionally write your own parameters in the walk function

	err = key.Walk(func(k *reggie.Key) error {
		fmt.Printf("Path: %s\n", k.Path)

		for name, value := range k.Values {
			fmt.Printf("- Key %s: Value: %v\n", name, value)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Walk failed: %v", err)
	}
}
```

Reggie aims to be as close to the original usage as possible while expanding on it. Let's take the usual way of creating a value inside a key with the `registry` pkg:

```go
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}

	err = key.SetBinaryValue("name", []byte("value"))
	if err != nil {
		log.Fatal(err)
	}

	err = key.SetQWordValue("name", 64)
	if err != nil {
		log.fatal(err)
	}
	... 
```
It can be cumbersome setting values like this. Instead, we have a helper function `CreateValue(...)`.
The function will infer the underlying type and process it accordingly. As a result, types can be custom:

```go
type QWORD uint64

func main() {
	reggieDemo, err := reggie.OpenKey(registry.CURRENT_USER, `Software\ReggieDemo`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}
	defer reggieDemo.Close()

	var val QWORD

	val = 12345

	err = reggieDemo.CreateValue("string key", val)
	if err != nil {
		log.Fatal(err)
	}
}
```

See the [examples folder](reggie/examples/) for more

## Contributing & License

Contributions are absolutely welcome. There is definitely improvements to be made and bugs unknown at present, so if you wish to contribute please check the [contributing](CONTRIBUTING.MD) markdown.

License is MIT, so you can take this and do whatever you want with it!
# reggie
Reggie is a wrapper over Golang's std `sys/registry` package.

This dev branch contains code for a stable v1 release, containing breaking but more permanent changes.

So far:

- `FillKeysValues()` is now `Load()` with better logic and possibility to limit how much is loaded.
    - Further development introduced `LoadWithLimit(limit int)` and `DeepLoad()`
- The structs for storing registry data is now less confusing and more straight cut
- Far less verbosity across the codebase
- `Traverse()` is now `Walk()` with improved and concise logic
- `ExportJson()` is now available to export registry key objects to JSON
- Functions relating to keys and values moved to their own `.go` files
- `DeleteValue()` and `GetValueAndNames()` new functions
- `CloneKey()` performs a full, deep in memory copy of a specified key to assign to a new object
- Updated to go version 1.24 and introduce usage of new `map` functions

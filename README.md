# reggie
Reggie is a wrapper over Golang's std `sys/registry` package.

This dev branch contains code for a stable v1 release, containing breaking but more permanent changes.

So far:

- `FillKeysValues()` is now `Load()` with better logic and possibility to limit how much is loaded.
- The structs for storing registry data is now less confusing and more straight cut
- Far less verbosity across the codebase
- `Traverse()` is now `Walk()` with better logic
... And more

v1 release roadmap will have the following:

[ ] - Export key datasets to JSON

[ ] - New helper functions such as `MustOpen()` or `MustCreate()` for quick testing / scripts

[ ] - Key / Value searching

More ideas soon to come.
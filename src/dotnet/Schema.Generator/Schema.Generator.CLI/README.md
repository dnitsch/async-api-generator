# Schema Generator CLI

The SchemaGenerator is a dotnet tool to inspect a built DLL (Assembly) and extract all the exported types, optionally supplying a namespace filter.

```
dotnet build 
```

```sh
bin/Debug/net7.0/schemagenerator
```

## Notes

> all the required DLLs must be present in the current directory of the inspected DLL, IF they are not you must copy them into that directory for the assembly loading process to work properly

Then run the downloaded/compiled binary for your architecture:

`schemagenerator -p path/to/Dll.dll -f Filter.This.Namespace.Only -o output/dir`

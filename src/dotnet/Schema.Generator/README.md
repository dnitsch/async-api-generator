# SchemaGenerator

Project to allow for schema generation from a "published" DLL.

The SchemaGenerator is a dotnet tool to inspect a built DLL (Assembly) and extract all the exported types, optionally supplying a namespace filter.

This generator uses [this project](https://github.com/RicoSuter/NJsonSchema) under the hood and furthermore use the System.Text.Json package by default - currently unable to override this. 

For details about System.Text.Json see [this part](https://github.com/RicoSuter/NJsonSchema/wiki/JsonSchemaGenerator#systemtextjson-vs-newtonsoftjson) of the documentation.

To get the most reliable schema output use the annotations to their fullest including setting examples in the XML tags which can then be exported/used by setting the [following property](https://github.com/RicoSuter/NJsonSchema/wiki/XML-Documentation#define-examples) in the csproj file. 

[Attribute Definitions](https://github.com/RicoSuter/NJsonSchema/wiki/JsonSchemaGenerator#provided-njsonschema-attributes)

## CLI

To see the list of all options run

```sh
schemagenerator --help
```

### Build locally

In the root of the repo run the following command to generate the binary.

```sh
make NETCORE_TARGET_ARCH=win-x64|osx-x64|osx-arm64 build_schemagen
./dist/[osx-x64|win-x64]/schemagenerator --help
```

### Download

Follow instructions for:

- [Windows]()

- [Intel Mac]()

- [ARM Mac]()

### Example

`schemagenerator -p path/to/publish/Dll.dll -f Filter.This.Namespace.Only -d -o output/dir`

> when running locally against a target - run `dotnet publish` in the target project to ensure all the dependency DLLs are present in the same directory, else assemblyLoader will fail if types are referencing packages from other assemblies.

Actual example against the `domain.Dft.Sorter6.Mapper` repo

First publish the `domain.Dft.Sorter6.Mapper.Models` project:

`dotnet publish /your/git/path/to/domain.Dft.Sorter6.Mapper/src/domain.Dft.Sorter6.Mapper.Models/domain.Dft.Sorter6.Mapper.Models.csproj -c Release`

Once done we can then generate the output.

`schemagenerator -p /your/git/path/to/domain.Dft.Sorter6.Mapper/src/domain.Dft.Sorter6.Mapper.Models/bin/Release/net6.0/publish/domain.Dft.Sorter6.Mapper.Models.dll -f domain.Dft.Sorter6.Mapper.Models.WarehouseEvents.Version1 -d`

By default the output is generated in the current directory creating a folder `.autogened` which will include both the samples and schemas.

> NOTE: Samples are using the null type on a complex object type as opposed the typed object.


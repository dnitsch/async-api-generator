using System.Text.Json;
using NJsonSchema.Generation;
using System.Reflection;
using NJsonSchema;
using System.Text.Json.Serialization;

namespace schemagenerator;

/// <summary>
/// Generate takes a list Namespace.Class+OptionalSubType and converts it to
/// <System.Type> instance to pass to the NJsonSchema package
/// </summary>
/// <remarks>
/// This shouldn't really live here - only for illustrative purposes
/// </remarks>
public class Generate(string dllPath, string[] namespaceFilter, bool deref, string outputDir)
{
    private readonly string _dllPath = dllPath;

    private string[] _namespaceFilter = namespaceFilter;

    private string _outDir = outputDir;
    private bool _derefSchema = deref;

    private string errorMsgs = "";

    private protected IList<Type> foundTypes { get; private set; } = [];

    /// <summary>
    /// Generates Schema from Types
    /// </summary>
    /// <returns></returns>
    /// <exception cref="Exception"></exception>
    public void SchemaFromTypes()
    {
        // MetadataLoadContext ctx = null;
        // try {
        ConvertToSystemType();
        // var ctx = ConvertToSystemType();

        var serializerOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
            AllowTrailingCommas = true,
            Converters = {
                    new JsonStringEnumConverter(JsonNamingPolicy.CamelCase, true)
                },
        };

        var settings = new JsonSchemaGeneratorSettings
        {
            SerializerSettings = null, // serializerSettings,
            FlattenInheritanceHierarchy = true, // try to inline all properties
            AllowReferencesWithProperties = false, // try to inline all properties
            SerializerOptions = serializerOptions, // use System.Text.Json
            SchemaType = NJsonSchema.SchemaType.JsonSchema,
            GenerateAbstractSchemas = false,
            GenerateCustomNullableProperties = false,
        };

        var generator = new JsonSchemaGenerator(settings);

        Directory.CreateDirectory($"{_outDir}/schemas");
        Directory.CreateDirectory($"{_outDir}/samples");

        Console.WriteLine($"found {foundTypes.Count} schemas");

        var tasks = new List<Task>();

        foreach (var systemType in foundTypes)
        {
            if (systemType.Namespace != null && systemType.FullName != null)
            {
                if (systemType.FullName.Contains('+'))
                {
#if DEBUG
                    Console.WriteLine($"INFO: Skipping: {systemType.FullName} - ONLY want TopLevel POCOs");
#endif
                    continue;
                }
                string outputName = (systemType.FullName)[(systemType.Namespace.Length + 1)..];

                var originalSchema = generator.Generate(systemType);

                tryWriteSchema(outputName, originalSchema, systemType.FullName, _derefSchema);
                tryWriteSample(outputName, originalSchema, systemType.FullName);
            }
        }
        if (errorMsgs.Length > 0)
            throw new Exception(errorMsgs);
    }

    private void tryWriteSchema(string outputName, JsonSchema schema, string typeName, bool deref)
    {
        try
        {
            File.WriteAllText($"{_outDir}/schemas/{outputName}.json", schema.ToJson());

            if (deref)
            {
                Helper helper = new Helper(new JsonSerializerOptions { WriteIndented = true });

                using var document = JsonDocument.Parse(schema.ToJson());
                var rootElement = document.RootElement;
                // pass in the definitions 
                var inlinedSchema = helper.RemoveReferences(rootElement, document);

                File.WriteAllText($"{_outDir}/schemas/{outputName}.schema.json", inlinedSchema.RootElement.GetRawText());
            }
        }
        catch (Exception ex)
        {
            var err = $"ERROR: Cannot write Schema for {outputName} in {typeName}, with err: {ex.Message}\n{ex.StackTrace}\n";
#if DEBUG
            Console.WriteLine(err);
#endif
            errorMsgs += err;
        }
    }

    private void tryWriteSample(string outputName, JsonSchema schema, string typeName)
    {
        try
        {
            File.WriteAllText($"{_outDir}/samples/{outputName}.sample.json", schema.ToSampleJson().ToString());
        }
        catch (Exception ex)
        {
            var err = $"ERROR: Cannot write Sample for {outputName} in {typeName}, with err: {ex.Message}\n{ex.StackTrace}\n";
#if DEBUG
            Console.WriteLine(err);
#endif
            errorMsgs += err;
        }
    }

    /// <summary>
    /// Attempts to convert strings to System.Type
    /// </summary>
    /// <remarks>
    /// </remarks>
    /// <param name="stringTypes"></param>
    /// <returns></returns>
    private void ConvertToSystemType()
    {
        try
        {
            var loadContext = new PluginLoadContext(_dllPath);
            var assembly = loadContext.LoadFromAssemblyName(new AssemblyName(Path.GetFileNameWithoutExtension(_dllPath)));
            foreach (var a in assembly.GetTypes())
            {
                if (_namespaceFilter.Any((f) => a.FullName.Contains(f)))
                {
                    Console.WriteLine($"INFO: adding type: {a.Name}, from path: {_dllPath}");
                    foundTypes.Add(a);
                }
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"ERROR: unable to dereference type from string: {_dllPath}, {ex.Message}\n{ex.StackTrace}");
        }
    }
}

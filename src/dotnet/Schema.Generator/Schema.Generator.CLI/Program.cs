using McMaster.Extensions.CommandLineUtils;
using schemagenerator;

var app = new CommandLineApplication();

app.HelpOption();

var path = app.Option("-p|--path <PATH_TO_DLL>", "The path to DLL", CommandOptionType.SingleValue);
var nsFilter = app.Option("-f|--filter <NAMESPACE>", "Namespace Filter, multiple can be provided", CommandOptionType.MultipleValue);
var derefSchema = app.Option("-d|--deref", "Dereference Schema. Removes addtional schema elements and inlines them in a single top level property", CommandOptionType.NoValue);
var output = app.Option("-o|--output <OUTPUT_DIR>", "Output directory", CommandOptionType.SingleValue);
output.DefaultValue = ".autogened";

var contextToken = new CancellationToken();

app.OnExecuteAsync((contextToken) =>
{
        var gen = new Generate(path.Value(), nsFilter?.Values.ToArray(), derefSchema.Values.Count > 0, output.Value());
        gen.SchemaFromTypes();
        return Task.FromResult(true);

});

return app.Execute(args);

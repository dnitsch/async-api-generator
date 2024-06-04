using McMaster.Extensions.CommandLineUtils;
using System.ComponentModel.DataAnnotations;


namespace schemagenerator
{
        [Command]
    public class Program
    {
        public static Task<int> Main(string[] args) => CommandLineApplication.ExecuteAsync<Program>(args);

        [Required]
        [Option(Description = "Path to the DLL", ShortName = "p")]
#pragma warning disable CS8618 // Non-nullable field must contain a non-null value when exiting constructor. Consider adding the 'required' modifier or declaring as nullable.
        public string DllPath { get; }
#pragma warning restore CS8618 // Non-nullable field must contain a non-null value when exiting constructor. Consider adding the 'required' modifier or declaring as nullable.

        [Required]
        [Option(Description = "Namespace Filter, multiple can be provided with a comma used as a separator", LongName = "namespace-filter", ShortName = "f", ShowInHelpText = true)]
#pragma warning disable CS8618 // Non-nullable field must contain a non-null value when exiting constructor. Consider adding the 'required' modifier or declaring as nullable.
        public string NamespaceFilter { get; }
#pragma warning restore CS8618 // Non-nullable field must contain a non-null value when exiting constructor. Consider adding the 'required' modifier or declaring as nullable.

        [Option(Description = "Output directory", LongName = "output-dir", ShortName = "o")]
        public string OutputDir { get; } = ".autogened";

        [Option(Description = "Dereference Schema", LongName = "deref-schema", ShortName = "d")]
        public bool Deref { get; } = false;

        private void OnExecuteAsync()
        {
            var gen = new Generate(DllPath, NamespaceFilter.Split(","), Deref, OutputDir); 
            gen.SchemaFromTypes();
        }
    }
}

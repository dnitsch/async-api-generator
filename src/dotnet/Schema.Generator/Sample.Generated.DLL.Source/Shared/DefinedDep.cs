using System.Text.Json.Serialization;

namespace Sample.Generated.DLL.Source.Shared
{
    public class DefinedDep
    {
        [JsonPropertyName("prop1")]
        public string Prop1 { get; set; }

        [JsonPropertyName("prop2")]
        public string Prop2 { get; set; }
    }
}


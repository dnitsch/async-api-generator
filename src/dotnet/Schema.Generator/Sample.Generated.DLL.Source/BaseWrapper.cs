using System.Text.Json.Serialization;

namespace Sample.Generated.DLL.Source
{

    public abstract class BaseWrapper
    {
        [JsonPropertyName("guid")]
        public string Guid { get; set; }

        [JsonPropertyName("creationDate")]
        public string CreationDate { get; set; }

        [JsonPropertyName("version")]
        public string Version => VersionConstants.NextVersion1;
    }

    public static class VersionConstants
    {
        public const string DftVersion1 = "v1";
        public const string NextVersion1 = "1";
    }
}


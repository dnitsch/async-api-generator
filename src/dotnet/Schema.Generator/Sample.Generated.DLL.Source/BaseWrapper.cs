using System.Text.Json.Serialization;

namespace Sample.Generated.DLL.Source
{

    public abstract class BaseWrapper
    {
        [JsonPropertyName("guid")]
        public string Guid { get; set; }

        [JsonPropertyName("creationDate")]
        public string CreationDate { get; set; }
    }


}


using System.Text.Json.Serialization;
using NJsonSchema.Annotations;
using Sample.Generated.DLL.Source.Shared;

namespace Sample.Generated.DLL.Source.TakeThisModel;

[JsonSchemaFlatten]
public class Another : BaseWrapper
{
    //public class BodyPayload
    //{
    //    [JsonPropertyName("orderId")]
    //    public string OrderId { get; set; }

    //    [JsonPropertyName("status")]
    //    public string Status { get; set; }

    //    [JsonPropertyName("isCancelled")]
    //    public bool IsCancelled { get; set; }

    //    // with items reference from another class
    //    [JsonPropertyName("items")]
    //    public IEnumerable<DefinedDep> Items { get; set; }
    //}

    //[JsonPropertyName("payload")]
    //public BodyPayload Payload { get; set; }
}

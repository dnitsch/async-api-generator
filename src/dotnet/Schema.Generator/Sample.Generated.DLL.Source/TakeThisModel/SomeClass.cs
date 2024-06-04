using System.Text.Json.Serialization;
using NJsonSchema.Annotations;

namespace Sample.Generated.DLL.Source.TakeThisModel;

public class SomeClass : BaseWrapper
{
    public class BodyPayload
    {
        [JsonPropertyName("orderId")]
        public string OrderId { get; set; }

        [JsonPropertyName("status")]
        public string Status { get; set; }

        [JsonPropertyName("isCancelled")]
        public bool IsCancelled { get; set; }

        // with nested items in the main payload
        [JsonPropertyName("items")]
        public IEnumerable<Item> Items { get; set; }

        public class Item
        {
            [JsonPropertyName("sku")]
            public string Sku { get; set; }

            [JsonPropertyName("itemReference")]
            public string ItemReference { get; set; }
        }

        [JsonPropertyName("payload")]
        public BodyPayload Payload { get; set; }
    }
}



using NJsonSchema.Annotations;
using System.Text.Json.Serialization;

namespace Sample.Generated.DLL.Source.DoNotTakeThisModel
{
    [JsonSchemaFlatten]
    public class SomeClass2 : BaseWrapper
    {
        [JsonPropertyName("payload")]
        public MessagePayload Payload { get; set; }

        public class MessagePayload
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
        }
    }

}
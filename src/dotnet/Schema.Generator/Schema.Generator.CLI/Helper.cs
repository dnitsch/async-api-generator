using System.Text.Json;

namespace schemagenerator;

public class Helper(JsonSerializerOptions options)
{
    private JsonSerializerOptions _options = options;

    public JsonDocument RemoveReferences(JsonElement schema, JsonDocument originalDocument)
    {
        if (schema.ValueKind.Equals(JsonValueKind.Array))
        {
            var updatedItems = new List<JsonElement>();
            foreach (var item in schema.EnumerateArray())
            {
                var derefedProp = isRefObject(item, originalDocument);
                if (derefedProp.DefFound) {
                    updatedItems.Add(RemoveReferences(derefedProp.Definition, originalDocument).RootElement);
                }
                else {
                    updatedItems.Add(RemoveReferences(item, originalDocument).RootElement);
                }
            }
            return JsonDocument.Parse(JsonSerializer.Serialize(updatedItems, this._options));
        }
        else if (
        schema.ValueKind.Equals(JsonValueKind.False) || 
        schema.ValueKind.Equals(JsonValueKind.True) ||
        schema.ValueKind.Equals(JsonValueKind.Number) ||
        schema.ValueKind.Equals(JsonValueKind.Null) ||
        schema.ValueKind.Equals(JsonValueKind.Undefined)) {
            return JsonDocument.Parse(JsonSerializer.Serialize(schema, this._options));
        }
        else if(schema.ValueKind.Equals(JsonValueKind.String)) {
            return JsonDocument.Parse(JsonSerializer.Serialize(schema, this._options));
        }
        else
        {
            var updatedProperties = new Dictionary<string, JsonElement>();
            foreach (var property in schema.EnumerateObject())
            {
                var derefedProp = isRefObject(property.Value, originalDocument);
                if (derefedProp.DefFound) {
                    updatedProperties[property.Name] = RemoveReferences(derefedProp.Definition, originalDocument).RootElement;
                } else {
                    updatedProperties[property.Name] = RemoveReferences(property.Value, originalDocument).RootElement;
                }
            }
            // This will remove all definitions property in an object at that level
            // may not be desireable
            updatedProperties.Remove("definitions");
            return JsonDocument.Parse(JsonSerializer.Serialize(updatedProperties, this._options));
        }
    }

    public class IsRefResponse {
        public bool DefFound {get; internal set; }
        public JsonElement Definition { get; internal set; }
    }

    private IsRefResponse isRefObject(JsonElement property, JsonDocument originalDocument) {
        IsRefResponse resp = new IsRefResponse{DefFound = false}; // is there a better way to do this without an allocation

        try {
            if (property.ValueKind.Equals(JsonValueKind.Object)) {
                bool isRef = property.TryGetProperty("$ref", out JsonElement refProperty);
                var definitionRef = refProperty.GetString();
                if (definitionRef != null && definitionRef.StartsWith("#/definitions/"))
                {
                    var definitionName = definitionRef.Substring("#/definitions/".Length);
                    if (originalDocument.RootElement.TryGetProperty("definitions", out var definitionsProperty) && definitionsProperty.ValueKind == JsonValueKind.Object)
                    {
                        if (definitionsProperty.TryGetProperty(definitionName, out var definition))
                        {
                            resp.DefFound = true;
                            resp.Definition = definition;
                        }
                    }
                }
            }
            return resp;
        }
        catch (InvalidOperationException)
        {
            return resp;
        }
    }
}

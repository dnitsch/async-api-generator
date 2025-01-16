# AsyncAPI Bindings

[AsyncAPI standard spec](https://www.asyncapi.com/docs/reference/specification/v2.6.0#asyncapi-specification) describes all the possible elements that a valid asyncAPI document can have. 

Currently not all bindings are supported.

## Required

The current [AsyncAPI standard spec](https://www.asyncapi.com/docs/reference/specification/v2.6.0) is at version `2.6.0`.

The tool will deal with all the relevant sections to be able to build an AsyncAPI spec file from within a single repo.

The asyncAPI is built from the `Application` - i.e. service down, each service will have a toplevel description - `info` key, which will in turn include `channels`.

### Channels

Channels is a map of Channel - where a channel is an entity describing the transport layer. This can be a ServiceBus Queue/Topic.

The name of the channel, whilst it may contain all sorts of unicode and whitespace characters, it SHOULD conform to a machine-readable standard as it will be processed multiple times during the generation process.

See [notes](./notes.md) for the relationship diagram and precedence hierarchy

### Annotateable Properties

|annotation key|required?|description|options|examples|
|---|---|---|---|---|
|`category`|yes|Which part of the AsyncAPI document will this snippet relate to|`["root","info","server","channel","operation","subOperation","pubOperation","message"]`||
|`type`|yes|The type of a propery in an AsyncAPI section |`["json_schema","example","description","title","nameId"]`||
|`id`|yes (except on root/info)| name of the service. Will default to parent folder name - unless overridden. will be converted to this format:`urn:$business_domain:$bounded_context_domain:$service_name` => `urn:domain:packing:domain.packing.app`|||
|`parent`|no|The parent of this annotation if a message or operation ||

### Examples

- `type`: Example

```cs
//+gendoc id=SomeEvent category=message type=example
namespace domain.Packing.Services.PackArea.Contracts.Events;

public class SomeEvent : domainMessage<SomeEventPayload>
{
    public SomeEvent(SomeEventPayload payload)
    {
        MessageTypeName = nameof(SomeEvent);
        SourceSystem = PackAreaServiceConstants.Name;
        Guid = Guid.NewGuid();
        CreationDate = DateTime.UtcNow;
        Number = 1;
        NumberOf = 1;
        Owner = string.Empty;
        Stream = String.Empty;
        Payload = payload;
    }
}
//-gendoc
```

- `type`: JSON_schema can be defined in any file like below

```cs
namespace domain.Packing.Services.PackArea.Contracts.Events;

public class Bar
{
    public string Type { get; set; }

    public string Name { get; set; }
}

/* below is an example of schema in code
//+gendoc id=SomeEvent category=message type=json_schema
{
    "$schema": "http://json-schema.org/draft-07/schema",
    "$id": "http://example.com/example.json",
    "type": "object",
    "required": [
        "payload",
        "guid",
        "creationDate",
        "messageTypeName",
        "version",
        "owner",
        "stream",
        "sourceSystem"
    ],
    "properties": {
        "payload": {
            "$id": "#/properties/payload",
            "type": "object",
            "required": [
                "packAreaId",
                "enabled",
                "warehouseCode",
                "mode",
                "itemThreshold",
    ...truncated for brevity
}
//-gendoc
*/

```

> However, the recommended way is to keep your schema in a file named => `MESSAGE_NAME.schema.json` [see example](../src/test/domain.sample/src/schemas/SomeEvent.schema.json)


## Nice To Have

- `servers` keyword describes the technology providing the transport layer - e.g. Kafka/RabbitMQ/AWS SQS/Azure ServiceBus
    - it may contain a map of multiple implementations - e.g. dev/preprod/prod 
    - > in conjunction with a channel key a full URL can be constructed to use by the client(s) to either publish or subscribe to messages on that ServiceBus's Topic/Queue/Topic-Subscription

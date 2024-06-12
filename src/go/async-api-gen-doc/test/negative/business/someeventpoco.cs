using System;

//+gendoc:Event type=message,consumer=sddsffsd,key=value,key1=value
namespace Foo.App.Contracts.Events;

public class SomeEvent : FooMessage<SomeEventPayload>
{
    public SomeEvent(SomeEventPayload payload)
    {
        MessageTypeName = nameof(SomeEvent);
        Guid = Guid.NewGuid();
    }
}

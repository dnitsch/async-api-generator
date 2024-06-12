using System;

/*
//+gendoc category=pubOperation type=description id=BuxQuz parent=foo-stuff~operation-cancelled-domain-event isSub=true
this operation publishes this message
//-gendoc
*/

// parent is the channel i.e. queue or topic name 
// id is the name of the message itself
//+gendoc id=BuxQuz category=message type=example
namespace Foo.Bar.Contracts.Events;

public class BuxQuz : FooMessage<BuxQuzPayload>
{
    public BuxQuz(BuxQuzPayload payload)
    {
        MessageTypeName = nameof(BuxQuz);
        Guid = Guid.NewGuid();
        ...
    }
}
//-gendoc

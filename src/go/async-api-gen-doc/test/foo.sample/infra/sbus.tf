locals {
  topics = [
    //+gendoc category=channel type=nameId service_id=[foo.bar.app]
    "foo-stuff~operation-cancelled-domain-event",
    //-gendoc
    "foo-stuff~operation-updated-domain-event",
    "foo-stuff~operation-completed-domain-event",
    "foo-stuff~operation-skipped-domain-event",
  ]

  subscriptions = [
    {
      //+gendoc category=channel id=foo-stuff-subscription service_id=[foo.bar.app]
      topic : azurerm_servicebus_topic.cancelled-topic,
      name : "foo-stuff-subscription"
      //-gendoc
    },
    {
/*
//+gendoc category=channel id=foo-stuff-subscription type=description service_id=[foo.bar.app]
This channel is used to convey subscriptions
//-gendoc
*/
      topic : azurerm_servicebus_topic.ItemSortation-waypointComplete-topic,
      //+gendoc category=channel id=foo-stuff-subscription type=nameId service_id=[foo.bar.app]
      //-gendoc
      name : "foo-stuff-subscription"
    },
    {
      topic : azurerm_servicebus_topic.ItemSortation-operationWaypointStarted-topic,
      name : "foo-stuff-subscription"
    },
   
  ]
}

resource "azurerm_servicebus_topic" "default" {
  for_each            = toset(local.topics)
  ...
}

# Add the proper subscriptions
resource "azurerm_servicebus_subscription" "default" {
  for_each            = toset(local.topics)
  ... 
}

# pub-sub 

_pub-sub_ is a simple Go library that allows you to implement a simple Publish-Subscribe mechanism, and a simple mechanism for joining asynchronous Subscribers together.  

It uses go-routines and channels, not queues.  

### Go version 
This library assumes Go version 1.3.3+. Previous versions of Go have a different way of approaching GOPATH, etc, hence Makefile would have to be done differently. 

## The API
 
The API exposes 4 interfaces and 2 functions:
+ Publisher -- which represents the Sender of events, and is a shorthand for func(interface{}). Invoking the Publisher - or invoking the function - is the act of sending of an event. 
+ Subscriber -- which represents the Receiver of events, and is a shorthand for func(interface{}). You do not invoke the Subscriber, but it is invoked for you when you are subscribing to a topic. The event is passed as the only parameter to the call. 
+ Topic -- which represents the typical Topic parties can subscribe to. Each Topic has a name, which in theory should identify it uniquely among other topics. The implementation does not 
use this field, if only for informative reasons. Topics allow you to create Publishers and Subscribers. Bear in mind, that since queues are not used, events are _blocked_ when you invoke Publishers, until at least one Subscriber is available. This is to prevent a situation where Publishing occurs before Subscribing.  
+ NewTopic -- is a public function that allows you to create a Topic with a name. 
+ NamedEvents -- which represents a map of events. As mentioned below, events are assumed to be represented by _interface{}_. A batch of events - from various topics - can be henceforth represented by a Go map, where each key reflects the name of the Topic. This construct is useful for the AwaitAll function. 
+ AwaitAll -- is a public function that allows you to subscribe to multiple Topics at once, and wait until all of them have been notified by a Publish. Hence it is a logical AND gate of multiple Topic subscriptions. Do note, that the function may wait indefinitely, if one of the Topics does not have a Publish. Still, the advantage in this method, is that it does not panic. 
+ MustAwaitAll -- is a public method that allows you to subscribe to multiple Topics at once, and wait until all of them have been notified by a Publish, or a specified duration of time lapses - whichever occurs earlier. Like, AwaitAll - this is a logical AND gate of multiple Topic subscriptions. Do note, that if the expected time lapses, and a Publish did not arrive at a specified Topic, this function will panic.   

An important assumption of the implementation is that an event is represented by _interface{}_. The framwork does not place any assumptions about type. 

### Simple example of using Publisher and Subscriber
```go
import (
    "github.com/tholowka/pub-sub/events"
    "log"
)

topic := events.NewTopic("my-new-topic")
publisher := topic.NewPublisher()
subscriber := func(event interface{}) {
   log.Println(event) 
}
topic.NewSubscriber(subscriber)

publisher("Inform about an event")
```

### Simple usage of the AwaitAll() functions (a Joiner pattern on an array of Topics)
```go
import (
    "github.com/tholowka/pub-sub/events"
    "log"
)

firstTopic := events.NewTopic("my-new-topic")
secondTopic := events.NewTopic("my-latest-topic")
publisher := topic.NewPublisher()
subscriber := func(event interface{}) {
   log.Println(event) 
}
topic.NewSubscriber(subscriber)

publisher("Inform about an event")
awaitForResult := AwaitAll([]Topic { firstTopic, secondTopic}, time.Duration(10)*time.Second)
result := <-awaitForResult
```

### A simple Dependency Injection framework
The above example shows something that can be seen as a simplified DI framework. Example provided below will try to clarify this a bit further.

In a traditional DI pattern, effort is made to 
+ separate high-level and low-level objects 
+ make sure that dependent objects obtain dependencies in a loosely coupled way
+ separate the creation of an object's dependencies from its own behaviour or creation

Alternatives exist to the provided approach. See http://blog.parse.com/2014/05/13/dependency-injection-with-go/

This library allows you to achieve Dependency Injection via some more-or-less clever channeling in the background. 

Consider the following three classes:
```go
type Address struct {
    street string
    postCode string
    country string
}

type PersonalDetails struct {
    firstName string
    lastName string
    birthDate string
}

type Customer struct {
    personalDetails *PersonalDetails
    address *Address
}
```

Clearly, _Customer_ functions as the high-level object, with _PersonalDetails_ and _Address_ being its dependencies/constituents.
In typical, non-DI way, _Customer_ would be created with some form of constructor injection (or setter usage):
```go
    customer := &Customer { personalDetails, address } 
```
As the code or complexity of the object rises, the above construction pattern becomes a pain to test and maintain. You are forced to create a chain of constructors (maintained by
the high-level object) in order to construct a high-level object from its low-level dependencies.  
Raising the number of dependencies often ends with the creation of mid-level objects, that become a burden to maintain, etci, because of the need to test them and make them interchangeable... 

You can proceed a bit differently than with using a chain of constructors (or setters). 

Consider a hypothetical Wire(...) method and a Configuration struct shown below. 

```go
const (
//  The const values below are arbitrary strings, yet I wanted to express the fact that this is mainly for singleton creation usage.
    ADDRESS = "address_singleton"
    PERSONAL_DETAILS = "personalDetails_singleton"
    CUSTOMER = "customer_singleton"
)
type Configuration struct {
    modules map[string]events.Topic
}

func (c *Configuration) Add(module events.Topic) {
    if _, exists := c.modules[module.String()]; exists {
        //do not allow of overwriting of modules.
        return
    }
    c.modules[module.String()] = module
}
func (c *Configuration) Get(name string) events.Topic {
    return c.modules[name]
}

func Wire(configuration *Configuration) {
    dependencies := <-events.MustAwaitAll([]events.Topic{
        configuration.Get(ADDRESS),
        conciguration.Get(PERSONAL_DETAILS)
    }, time.Duration(1)*time.Second)

    //The objects need to cast, because events.MustAwaitAll returns a map[string]interface{}
    //If Go had generics...
    address := dependencies[ADDRESS].(*Address)
    personalDetails := dependencies[PERSONAL_DETAILS].(*PersonalDetails)
    //create the Customer instance just like normal...
    customer := &Customer { personalDetails, address }
    //...but add it back to the Configuration
    newDependency := events.NewTopic(CUSTOMER) 
    configuration.Add(CUSTOMER, newDependency)
    //alert all high-level objects that might await for a Customer instance that it is now available
    newDependency.NewPublisher(customer)
}
```


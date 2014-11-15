# pub-sub 

_pub-sub_ is a simple Go library that allows you to:
+ use a Publish-Subscribe mechanism. 
+ allows you to join Subscribers in a Publish-Subscribe pattern together so that a piece of code executes when they all have been notified (known also as a Join pattern) 
+ allows you to construct a simplified DI engine with some clever coding on your part

Go-routines and channels are used in the background, not queues.  

### Go version 
This library assumes Go version 1.3.3+. Previous versions of Go have a different way of approaching GOPATH, etc, hence Makefile would have to be done differently. 

## The API
As in a traditional Publish-Subscriber pattern, this library provides a concept of a Topic (known as a Message bus elsewhere). Clients of this library construct a Topic, and 
then may assign Publishers and Subscribers to it in an (a)synchronous way. 
Publishers send the news, Subscribers react to it. 
 
The API exposes 4 interfaces and several functions:
+ _Publisher_ -- which represents the Sender of events, and is a shorthand for _func(interface{})_. Invoking the Publisher (or invoking the function) - is the act of sending of an event. 
+ _Subscriber_ -- which represents the Receiver of events, and is a shorthand for _func(interface{})_. You do not invoke the Subscriber, but it is invoked for you when you are subscribing to a Topic. 
The event that the Publisher sent is passed as the parameter to the function call. 
+ _Topic_ -- which represents the typical Pub-Sub _Topic_ parties can subscribe to. Each Topic has a name, which in theory should identify it uniquely among other topics. The implementation does not 
use this field, and if only - it's for informative reasons. Topics allow you to create Publishers and Subscribers. Bear in mind: since queues are not used, events are _blocked_ when you invoke 
Publishers, until at least one Subscriber is available. This is to prevent a situation where Publishing occurs before Subscribing.  
+ _NewTopic_ -- is a public function that allows you to create a Topic with a name. 
+ _NamedEvents_ -- which represents a map of events. As mentioned below, events are assumed to be represented by _interface{}_. A batch of events - from various topics - can 
be henceforth represented by a Go map, where each key reflects the name of the Topic. This construct is useful for the _AwaitAll/MustAwaitAll_ function. 
+ _AwaitAll_ -- is a public function that allows you to subscribe to multiple Topics at once, and wait until all of them have been notified by a Publish. Hence it is a logical AND gate of 
multiple Topic subscriptions. Do note, that the function may wait indefinitely, if one of the Topics does not have a Publish. Still, the advantage in this method, is that it does not panic. 
+ _MustAwaitAll_ -- is a public function that allows you to subscribe to multiple Topics at once, and wait until all of them have been notified by a Publish, or a 
specified duration of time lapses - whichever occurs earlier. Like, AwaitAll - this is a logical AND gate of multiple Topic subscriptions. Do note, that if 
the expected time lapses, and a Publish did not arrive at a specified Topic, this function will panic.   

An important assumption of the implementation is that an event is represented by _interface{}_. The framework does not place any assumptions about type. 

### Simple example of using Publisher and Subscriber
Here I'll show how to create a Publisher and a Subscriber. 
The Publisher is the supplier of notification messages, the Subscriber receives them. 
```go
import (
    "github.com/tholowka/pub-sub/events"
    "log"
)

topic := events.NewTopic("my-new-topic")
publisher := topic.NewPublisher()
subscriber := func(event interface{}) {
    //prove to me that the Subscriber ran!
   log.Println(event) 
}
topic.NewSubscriber(subscriber)

publisher("Inform about an event")
```

### Simple usage of the AwaitAll() functions (a Join pattern on an array of Topics)
This example builds on the previous one and shows how to implement a Join pattern on the Topics. Imagine, you have more than one Topic, and you want to wait
until all have been notified. Here's how you may go at it:
```go
import (
    "github.com/tholowka/pub-sub/events"
    "log"
)

firstTopic := events.NewTopic("my-new-topic")
secondTopic := events.NewTopic("my-latest-topic")
publisher := topic.NewPublisher()
subscriber := func(event interface{}) {
    //prove to me that something was sent...
   log.Println(event) 
}
topic.NewSubscriber(subscriber)

publisher("Inform about an event")
awaitForResult := AwaitAll([]Topic { firstTopic, secondTopic}, 
    time.Duration(10)*time.Second)
result := <-awaitForResult
```
Do note the last line. In this line, you are waiting for a collection of results from all used Topics. This line will _block_ if 10 seconds pass (we are
using a 10 second timeout in the call to _AwaitAll_). This aspect depends on the function used: 
if you were to use _MustAwaitAll_ (instead of _AwaitAll_), the call to MustAwaitAll will _panic_ if Publish events do not 
occur on all Topics in specified time (hence, 
<-awaitForResult will either succeed or you have a _panic_). In case of the used _AwaitAll_ you do not 
receive a _panic_, but your call to <-awaitForAll will block. 

Personally, I prefer MustAwaitAll as it makes more sense to me, but there are people who do not like code that _panics_. 
Next version of this stack will most probably have _AwaitAll_ return an error in such cases (so it's in the spirit of 
standard Go libraries). 

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
Raising the number of dependencies often ends with the creation of mid-level objects, that become a burden to maintain, etc, 
because of the need to test them and make them interchangeable...If you sometimes wonder why you need all these unit-tests, then
you are not alone. 

However, you can proceed a bit differently here than just by using a chain of constructors (or setters).  

Consider a hypothetical Wire(...) method and a Configuration struct shown below. 

```go
const (
//The const values below are arbitrary strings, 
//yet I wanted to express the fact that this is 
//mainly for singleton creation usage.
//If you want to create numerous objects
//then this approach would probably be used to 
//make a factory
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
    //Running this asynchronously is important! 
    //You usually have - in most cases - a big graph of 
    //dependencies, where one object may create 
    //many others. That graph's 
    //execution tree might not necessarily 
    //reflect the order in which you call 
    //various Wire() 
    //methods in your stack. 
    go func () {
        dependencies := <-events.MustAwaitAll([]events.Topic{
            configuration.Get(ADDRESS),
            conciguration.Get(PERSONAL_DETAILS)
        }, time.Duration(1)*time.Second)

        //The objects need to cast, because 
        //events.MustAwaitAll returns a map[string]interface{}
        //If Go had generics...
        address := dependencies[ADDRESS].(*Address)
        personalDetails := dependencies[PERSONAL_DETAILS].(*PersonalDetails)
        //create the Customer instance just like normal...
        customer := &Customer { personalDetails, address }
        //...but add it back to the Configuration
        newDependency := events.NewTopic(CUSTOMER) 
        configuration.Add(CUSTOMER, newDependency)
        //alert all high-level objects that might 
        //await for a Customer instance that it is now available
        newDependency.NewPublisher()(customer) 
        //in above line, we construct the function and instantly call it...
    }()
}
```

Two very important notes:
+ Note that Topics _block_ if there are no Subscribers listening. That's why MustAwaitAll function is used, which _panics_ if the wiring of modules does not finish in a prescribed amount of time. 
The _panic's_ message will alert you which Topics have not been Published in the given timeframe. You can specify the timeframe manually (see above example). 
+ Remember that Topics need to be *Closed*. You are using go-routines in the background which should be released when the Topics are no longer in usage. A pattern which you can implement is to 
add a method to the Configuration which you invoke after the Wire() method(s), that listens on a given Topic. When notified, it closes all Topics in the Configuration. 

Some minor notes:
+ Somewhere in your code you should create a _Configuration_ instance, before it is passed to _Wire_, so that it is initialised with all possible topics. Otherwise, _Get_ calls will return 
nils and your stack won't work. This has been omitted in the above example, for clarity. I usually use one _Configuration_ instance for the whole app, and initialise it 
with const strings once somewhere close to a main() method.
+ I usually keep many _Wire(configuration *Configuration)_ functions in my stack: one per package or so. I call them in sequence, but due to the asynchronous architecture  of the Pub-Sub and the Wire() method, objects are created when they need to.  
+ The Add(...) method above *escapes* if a given Topic already exists in the map. This is important, and obvious if you think about it for a second: you might be wiping out 
somebody's Subscribers, if you allow overwriting. And by doing so, you are in strong risk of a _panic_ or at least an indefinite block/deadlock. 
+ The Configuration model you have above is not go-routine proof. However, in most cases the act of configuring of an app is something you run sequentially in a single 
go-routine. This might not be true always. If you are forced to use _Add()_ or _Get()_ from go-routines, you should change the implementation to be go-routine-proof: 
pass state via channel calls. Passing state via channels (known also as the 'share by communicating' pattern) has been stronly supported by the inventors of Go language, 
search for it and absolution shall be yours.  
+ The above framework relies on _strings_ to do the wiring. They are used as _Topic_ names, and can lead to problems if they are not unique. This is a shortcoming of this library, 
but then this is not a DI engine done by a bunch of people. On the other hand, contrary to the other DI engine I saw out there, this does not rely on property markup (reminiscent of the Go 
_json/encoding_ package) to do the wiring. While I see advantages of this approach (and find it a good solution in case of _json/encoding_ package), I find I prefer a more explicit 
DI approach, with a clear pattern of what is called and when. With the '`' notation, the wiring is hidden. If it works, it is beautiful. If it goes wrong, you are in for a search.
Also, I was just happy to do it with channels and go-routines, not via reflection. 

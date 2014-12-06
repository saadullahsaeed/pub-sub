# pub-sub 

_pub-sub_ is a simple Go library that allows you to:
+ use a Publish-Subscribe mechanism. 
+ allows you to join Subscribers in a Publish-Subscribe pattern together so that a piece of code executes when they have been notified (known also as a Join pattern) 
+ allows you to construct a simplified DI engine with some clever coding on your part

Go-routines and channels are used in the background, not queues.  

### Go version 
This library assumes Go version 1.3.3+. Previous versions of Go have a different way of approaching GOPATH, etc, hence Makefile would have to be done differently. 

## Usage ##
To use the latest 'master' branch version, invoke the following: 
```go
go get github.com/tholowka/pub-sub/events
```
and import the library using traditional means:
```go
import (
    "github.com/tholowka/pub-sub/events"
)
```

## The API
As in a traditional Publish-Subscriber pattern, this library provides a concept of a Topic (known as a Message bus elsewhere). Clients of this library construct a Topic, and 
then may assign Publishers and Subscribers to it in an (a)synchronous way. 
Publishers send the news, Subscribers react to it. 

The other part of the library is support for a Join pattern. Two operations are provided: an AND and an OR (you can think of signal processing here, as in a circuit). 
The AND gate returns a map of all Published events or terminates with an error. The OR gate returns a map of a Published (out of many) or terminates with an error. 
Two additional helper methods are provided, preceded with '_Must_': they replicate the base behaviour (of functions without the _Must_), but _panic_ in case of errors. 
This is useful in code, that if wrong, should terminate the application.
 
The API exposes 4 interfaces and several functions:
+ _Publisher_ -- which represents the Sender of events, and is a shorthand for _func(interface{})_. Invoking the Publisher (or invoking the function) - is the act of sending of an event. 
+ _Subscriber_ -- which represents the Receiver of events, and is a shorthand for _func(interface{})_. You do not invoke the Subscriber, but it is invoked for you when you are subscribing to a Topic. 
The event that the Publisher sent is passed as the parameter to the function call. 
+ _Topic_ -- which represents the typical Pub-Sub _Topic_ parties can subscribe to. Each Topic has a name, which in theory should identify it uniquely among other topics. The implementation does not 
use this field, and if only - it's for informative reasons. Topics allow you to create Publishers and Subscribers. Bear in mind: since queues are not used, events are _blocked_ when you invoke 
Publishers, until at least one Subscriber is available. This is to prevent a situation where Publishing occurs before Subscribing.  
+ _NewTopic_ -- is a public function that allows you to create a Topic with a name. 
+ _And_ -- is a public function that allows you to subscribe to multiple Topics at once, and wait until all of them have been notified by a Publish. 
Hence it is a logical AND gate of multiple Topic subscriptions. Since Publish events might occur repeatedly on one of the provided Topics before data is passed to the returned Topic, 
the actual type of the returned data is _map[string][]interface{}_, where each key of the map reflects the name of one of the provided Topics.  
+ _Or_ -- is a public function that allows you to subscribe to multiple Topics at once, and wait until ANY of them has been notified by a Publish. 
Hence it is a logical OR gate of multiple Topic subscriptions. Just like And, the returned type is _map[string][]interface{}_. The reason is that a common function/pattern is used behind the scenes, 
it also promotes a uniform interface.  
+ _WhenTimeout_ -- is a public function that allows you to construct a side Topic, which publishes messages whenever a base provided Topic is not
updated with an event in a prescribed amount of time.
+ _MustPublishWithin_ -- is very similar to _WhenTimeout_ except that it _panics_ when the underlying Topic picks up a time.Time event.
+ _NewTemporaryTopic_ -- is a public function which allows you to construct a Topic, that is automatically closed after a certain period of time. This functionality is useful if you don't want to worry about the lifecycle of the Topic. The downside is that you can't reuse a closed Topic. 
+ _NewTimerTopic_ -- is a public function that allows you to construct a special version of a Topic, which encapsulates over a Go time.NewTicker(). 
Any Publish event occuring on the Timer resets the Timer. This Topic implementation is used within _WhenTimeout_

Some of the factories in this library have a _WithLogging_ variant, which allows you to inspect Topic state via a logging method (like log.Println). This
has a performance impact of course.  

An important assumption of the implementation is that an event is represented by _interface{}_. The framework does not place any assumptions about type. 

Note, that the library exposes a Version() method which you can use to inspect this libraries' version.  

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
topic.Close()
```

### Simple usage of the And() functions (a Join pattern on an array of Topics)
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
firstTopic.NewSubscriber(subscriber) 
secondTopic.NewSubscriber(subscriber)

publisher("Inform about an event")
topic := And([]Topic { firstTopic, secondTopic}, "my-latest-rants")
firstTopic.Close()
secondTopic.Close()
```

What's probably worthy of mentioning at this point, is that AND collects all data before it executes. Hence, if any of your topics Publish to the same Topic repeatedly, 
before And fires, all of that data is preserved and provided to you in the callback. That's one of the reasons why, the actual type of the passed data is 
_map[string][]interface{}_.
Apart from the above comment, usage of OR is identical to the example above. The Or function is actually simpler, since data does not have to be collected. 
Still, the backing code converts the result into the same kind of structure (_map[string][]interface{}_), so that your event handling code for AND and OR results can be reused. 


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
        topic := events.And([]events.Topic{
            configuration.Get(ADDRESS),
            configuration.Get(PERSONAL_DETAILS)
        })
        topic.NewSubscriber(func(rawResult interface{}) {
            //The objects need to cast... If Go had generics...
            //As mentioned above, AND returns map[string][]interface{} structs
            result := rawResult.(map[string][]interface{})
            address := result[ADDRESS][0].(*Address)
            personalDetails := result[PERSONAL_DETAILS][0].(*PersonalDetails)
            //create the Customer instance just like normal...
            customer := &Customer { personalDetails, address }
            //...but add it back to the Configuration
            newDependency := events.NewTopic(CUSTOMER) 
            configuration.Add(CUSTOMER, newDependency)
            //alert all high-level objects that might 
            //await for a Customer instance that it is now available
            newDependency.NewPublisher()(customer) 
            //in above line, we construct the function and instantly call it...
        })
    }()
}
```

Two very important notes:
+ Note that Topics _block_ if there are no Subscribers listening. Topics should not deadlock, as the crucial parts run - in the provided implementation - in 
separate go-routines. You could use a WhenTimeout() or MustPublishWithin wrapper on the returned Topic to be safe, if you suspect you might not be notified via a Publisher. 
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

### TODOs
+ Splitter functionality

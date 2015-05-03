# pub-sub 

_pub-sub_ is a simple Go library that allows you to:
+ use a Publish-Subscribe mechanism. 
+ allows you to create Gates, like AND/OR to join Subscribers in a Publish-Subscribe pattern together  

Go-routines and channels are used in the background, not queues.  

[![views](https://sourcegraph.com/api/repos/github.com/tholowka/pub-sub/.counters/views.svg)](https://sourcegraph.com/github.com/tholowka/pub-sub)

### Build status  

[![Build status](https://travis-ci.org/tholowka/pub-sub.svg?branch=master)](https://travis-ci.org/tholowka/pub-sub.svg?branch=master)

### Go version 
This library assumes Go version 1.3.3+. Previous versions of Go have a different way of approaching GOPATH, etc, hence Makefile would have to be done differently. 

### Documentation 

[![GoDoc](https://godoc.org/github.com/tholowka/pub-sub/events?status.svg)](https://godoc.org/github.com/tholowka/pub-sub/events)

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
 
The API exposes 4 interfaces and a function:
+ _Publisher_ -- which represents the Sender of events, and is a shorthand for _func(interface{})_. Invoking the Publisher (or invoking the function) - is the act of sending of an event. 
+ _Subscriber_ -- which represents the Receiver of events, and is a shorthand for _func(interface{})_. You do not invoke the Subscriber, but it is invoked for you when you are subscribing to a Topic. 
The event that the Publisher sent is passed as the parameter to the function call. 
+ _Topic_ -- which represents the typical Pub-Sub _Topic_ parties can subscribe to. Each Topic has a name, which in theory should identify it uniquely among other topics. The implementation does not 
use this field, and if only - it's for informative reasons. Topics allow you to create Publishers and Subscribers. Bear in mind: since queues are not used, events are _blocked_ when you invoke 
Publishers, until at least one Subscriber is available. This is to prevent a situation where Publishing occurs before Subscribing.  
+ _NewFactory_ -- is the public access point function that allows you to use this library. 

As such the _NewFactory_ method exposes you a _Factory_ interface providing the following methods:
+ _NewTopic_ -- is a public function that allows you to create a Topic with a name. 
+ _NewTickerTopic_ -- is a public function that allows you to construct a special version of a Topic, which encapsulates over a Go time.NewTicker(). 
+ _AndGate_ -- is a public function that allows you to subscribe to multiple Topics at once, and wait until all of them have been notified by a Publish. 
Hence it is a logical AND gate of multiple Topic subscriptions. Since Publish events might occur repeatedly on one of the provided Topics before data is passed to the returned Topic, 
the actual type of the returned data is _map[string][]interface{}_, where each key of the map reflects the name of one of the provided Topics.  
+ _OrGate_ -- is a public function that allows you to subscribe to multiple Topics at once, and wait until ANY of them has been notified by a Publish. 
Hence it is a logical OR gate of multiple Topic subscriptions. Just like with And, the returned type is _map[string][]interface{}_. 

An important assumption of the implementation is that an event is represented by _interface{}_. The framework does not place any assumptions about type. 
The actual type can actually vary, in some cases it is exactly what a Publish event has produced, in other cases -- see AndGate and OrGate -- it is actually an 
aggregation of such events.

Note, that the library exposes a Version() method which you can use to inspect this libraries' version.  

### Simple example of using Publisher and Subscriber
Here I'll show how to create a Publisher and a Subscriber. 
The Publisher is the supplier of notification messages, the Subscriber receives them. 
```go
import (
    "github.com/tholowka/pub-sub/events"
    "log"
)

topic := events.NewFactory().NewTopic("my-new-topic")
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

factory := events.NewFactory()
firstTopic := factory.NewTopic("my-new-topic")
secondTopic := factory.NewTopic("my-latest-topic")
publisher := topic.NewPublisher()
subscriber := func(event interface{}) {
    //prove to me that something was sent...
   log.Println(event) 
}
firstTopic.NewSubscriber(subscriber) 
secondTopic.NewSubscriber(subscriber)

publisher("Inform about an event")
topic := factory.AndGate([]Topic { firstTopic, secondTopic})
firstTopic.Close()
secondTopic.Close()
```

What's probably worthy of mentioning at this point, is that AND collects all data before it executes. Hence, if any of your topics Publish to the same Topic repeatedly, 
before And fires, all of that data is preserved and provided to you in the callback. That's one of the reasons why, the actual type of the passed data is 
_map[string][]interface{}_.
Apart from the above comment, usage of OR is identical to the example above. The Or function is actually simpler, since data does not have to be collected. 
Still, the backing code converts the result into the same kind of structure (_map[string][]interface{}_), so that your event handling code for AND and OR results can be reused. 

## Technical considerations 

The implementation provided by this library has changed between version 1.3 and 2.
Currently, each _Factory_ which allows you to build, join Topics, is backed by a single go-routine, and a number of channels. A _Factory_ has state: subscribers and topics created by it. Creation or closing of Topics results in a change of state, and hence has an impact on the overall performance of the library. 
Additionally, the implementations of Topics provided by this _Factory_ make sure that the act of Publishing (via NewPublisher()) or Subscribing (via NewSubscriber()) occurs in separate go-routines. Those go-routines are short-lived and terminate after the event is published or handled. 

## Benchmarks 

In the simplest scenario (one consumer, one producer) on a high-end Macbook (i7, 16GB) the result is 2500-3000 ns per operation. You can run the tests on your own system, via 'make a-benchmark-check' command. 
These tests also show that reusing a Publisher saves you some hundreds of nanoseconds. 

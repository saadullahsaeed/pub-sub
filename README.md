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


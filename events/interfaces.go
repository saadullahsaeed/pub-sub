/*
Provides a simple, 'Pub-Sub' library for Go. For a description of Publish-Subscriber (also known as Pub-Sub), please see 
See http://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern for a description of the Publish-Subscribe pattern.

Please note, that go routines and channels are used in the background, not queues for the purpose of delivering messages. This design has been recently 
revamped to use a single go-routine behind the scenes for routing messages. Each Subscribe and Publish event occurs in its own dedicated go-routine, so that 
in case it blocks - other go-routines are unaffected. Bear in mind the architecture here is highly asynchronous. 

Effort has been placed to make sure 'listening' occurs before an event is actually sent. In case of problems (which you can't cross out entirely due to the 
asynchronous behaviour) events that have been published but that do not find a recipient are 'requeued' back with a progressively bigger delay. 

For more information, please see the README at: https://github.com/tholowka/pub-sub.
*/
package events

import (
    "time"
)

/*
The Publisher in the Publish-Subscribe pattern, or a shorthand for a function which you may call each time you want to inform about a particular event. Publishers can be created by any Topic instance. 
*/
type Publisher func(interface{})
/*
The Subscriber in the Publish-Subscribe pattern, or a shorthand for a function which you may provide, that is invoked whenever an event is published to the Topic. 
*/
type Subscriber func(interface{})
/*
A typical Topic used in a Pub-Sub pattern. The Topic has a name, which in theory should identify it uniquely among other topics. 
The implementation does not use this name, unless for informative reasons. Topics can create Publishers and Subscribers.  
Topics can be closed and this closes the Topic permanently.
*/
type Topic interface {
    //Allows you to create a new Publisher for a Topic. If you provide a callback 
	//(optional) here, it is guaranteed to be invoked during the publishing act. 
	//Depending on the implementation it may be also used in a defer() section 
	//or called when the the publish failed. This may happen if Publishing is 
	//using on a Closed Topic. Publishing may occur in its own go-routine, and 
	//it's not guaranteed that the order you call Publishers is preserved 
	//(especially if you write to multiple Topics). 
    NewPublisher() Publisher
    //Allows you to register an arbitrary Subscriber for events in the Topic
    //Subscribing may occur in its own go-routine, hence even if the act of 
	//subscribing 'blocks' (for example due to the waiting on channel), the 
	//remainders of the Topic still execute normally. You can use the returned 
	//channel to await for the event, that the underlying Topic has picked up 
	//the Subscriber. This pattern may be important in applications where a 
	//Publish/Subscribe order is important, and you want to make sure the Publishing 
	//occurs after Subscribers have been picked up. 
    NewSubscriber(subscriber Subscriber)
    //Returns the topic's name
    String() string
    //Close frees the underlying resources, and depending on the implementation 
	//may render the Topic unusable
    Close() error
}

/**
This is the access point to the library. 
Exposes the top most layer of this library, which allows you to create Topics and Join them. 

The advantage of using Subscribers in factory methods (e.g. NewTopic/AndGate/OrGate) instead of calling them explicitly after Topic construction (i.e. NewSubscriber()) is that you don't get into a 
race condition: the subscribers are registered exactly at the moment the topic is registered.

Since 2.0
*/

type Factory interface {
	//Creates a new standard Topic
    NewTopic(string, ...Subscriber) Topic
	//Creates a Topic, backed by a Go Ticker, which can be subscribed
	//to for Tick events. Publishing to it does not make sense. 
    NewTickerTopic(string, time.Duration) Topic
	//Closes all Topics created by this Factory.
    Close() error
	//Creates a Topic implementing an AND gate (i.e. collecting
	//multiple events from various topics together and firing 
	//only when all Topic have been Published to)
    AndGate([]Topic, ...Subscriber) Topic
	//Creates a Topic implementing an OR gate (i.e. collecting
	//any event from various topics, and firing when any 
	//such event has been registered
    OrGate([]Topic, ...Subscriber) Topic
}

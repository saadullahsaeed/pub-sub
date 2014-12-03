package events

import (
    "time"
    "fmt"
    // "errors"
)

func NewTimerTopic(topicName string, timeout time.Duration) Topic {
    bus := &timeouter {
        topic {
            topicSpec {
                make(chan Subscriber),
                topicName,
                make(chan interface{}),
                make(chan bool),
                []Subscriber{},
                nil,
            },
        },
        timeout,
    }
    runTimeoutingTopicGoRoutine(bus.internal.newSubscribers,
        bus.internal.name,
        bus.internal.events,
        bus.internal.finish,
        bus.internal.subscribers,
        bus.internal.logger,
        timeout)
    return bus

}

type timeouter struct {
    internal topic
    timeout time.Duration
}

func (t *timeouter) NewPublisher() Publisher {
    return t.internal.NewPublisher()
}

func (t *timeouter) NewSubscriber(subscriber Subscriber) <-chan bool {
    return t.internal.NewSubscriber(subscriber)
}

func (t *timeouter) String() string {
    return t.internal.String()
}

func (t *timeouter) Close() error {
    return t.internal.Close()
}
/**
This function lies at the core of any Topic.

Essentially this is a state-machine, in which the provided channels change the state.
Starting process should await on the returned channel so that possibility of a deadlock is reduced (for example: an event added via
a Publisher and the subscribers have not been set -- a common scenario)
*/
func runTimeoutingTopicGoRoutine(newSubscribers <-chan Subscriber,
        name string,
        events <-chan interface{},
        finish <-chan bool,
        subscribers []Subscriber,
        logger func(...interface{}),
        timeout time.Duration) {

    go func() {
        closed := false
        timer := time.After(timeout)
        //note: line below is to make sure that subscribing occurs before ANY event publishing.
        if len(subscribers) == 0 {
            subscribers = append(subscribers, <-newSubscribers)
        }
        for ;; {
            if closed {
                if logger != nil {
                    logger(fmt.Sprintf("%v closed.", name))
                }
                return
            }
            select {
            case <-finish: //released when you close the channel
                closed = true
                return
            case newSubscriber:=<-newSubscribers:
                if closed {
                    if logger != nil {
                        logger(fmt.Sprintf("%v closed.", name))
                    }
                    return
                }
                //note: when channel is closed newSubscriber == nil
                if newSubscriber != nil {
                    subscribers = append(subscribers, newSubscriber)
                }
            case current:=<-timer:
                if closed {
                    if logger != nil {
                        logger(fmt.Sprintf("%v closed.", name))
                    }
                    return
                }
                //note: when channel is closed event == nil
                if logger != nil {
                    logger(fmt.Sprintf("%v notifying %v subscribers about timeout (%v).", name, len(subscribers), current))
                }
                for _, subscriber := range subscribers {
                    //note: if subscriber sends something to a channel we don't want to be blocked.
                    go subscriber(current)
                }
            case <-events:
                if closed {
                    if logger != nil {
                        logger(fmt.Sprintf("%v closed.", name))
                    }
                    return
                }
                timer = time.After(timeout)//resets timer
            }
        }
    }()
}
/**
Allows you to put an artificial timeout on a Topic, and send errors to a designated Topic whenever an event does not arrive in 
a specified amount of time. 

The created topic publishes events which are in fact errors. 
*/
func WhenTimeout(topic Topic, timeout time.Duration, timeoutTopicName string) Topic {
    timerTopic := NewTimerTopic(timeoutTopicName, timeout)
    timingPublisher := timerTopic.NewPublisher()
    subscriber := func(event interface{}) {
        timingPublisher(event)
    }
    topic.NewSubscriber(subscriber)
    return timerTopic
    // events := make(chan interface{})
    // closeChannel := make(chan bool)
    // timeouts := &timeoutingTopic { NewTopic(timeoutTopicName), events, closeChannel }
    // publisher := timeouts.NewPublisher()
    //
    // subscriber := func(event interface{}) {
    //     events<-event
    // }
    // topic.NewSubscriber(subscriber)
    // andListen := func() {
    //     var (
    //         timeoutChan <-chan time.Time
    //     )
    //     timeoutChan = time.After(timeout)
    //     for ;; {
    //         select {
    //         case <-closeChannel:
    //             return
    //         case <-events:
    //             timeoutChan = time.After(timeout)
    //         case <-timeoutChan:
    //             publisher(errors.New("Timeout on "+topic.String()))
    //         }
    //     }
    // }
    // go andListen()
    // return timeouts
}

type timeoutingTopic struct {
    topic Topic
    channel chan interface{}
    closeChannel chan bool
}

func (t *timeoutingTopic) NewPublisher() Publisher {
    return t.topic.NewPublisher()
}

func (t *timeoutingTopic) NewSubscriber(subscriber Subscriber) <-chan bool {
    return t.topic.NewSubscriber(subscriber)
}

func (t *timeoutingTopic) String() string {
    return t.topic.String()
}

func (t *timeoutingTopic) Close() error {
    close(t.channel)
    close(t.closeChannel)
    return t.topic.Close()
}

/**
This a variant of WhenTimeout, which panics instead of sending errors on a Topic
*/
func MustPublishWithin(topic Topic, timeout time.Duration) {
    errorTopic := WhenTimeout(topic, timeout, topic.String()+"-timeout-errors-collector")
    errorTopic.NewSubscriber(func(err interface{}) {
        go errorTopic.Close()
        panic(err.(error))
    })
}

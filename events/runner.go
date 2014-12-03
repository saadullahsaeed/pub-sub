package events

import (
    "fmt"
)

/**
This function lies at the core of any Topic.

Essentially this is a state-machine, in which the provided channels change the state.
Starting process should await on the returned channel so that possibility of a deadlock is reduced (for example: an event added via
a Publisher and the subscribers have not been set -- a common scenario)
*/
func runTopicGoRoutine(newSubscribers chan Subscriber,
        name string,
        events chan interface{},
        finish chan bool,
        subscribers []Subscriber,
        logger func(...interface{})) <-chan bool {

    topicIsReady := make(chan bool)
    go func() {
        closed := false
        //note: line below is to make sure that subscribing occurs before ANY event publishing.
        if len(subscribers) == 0 {
            subscribers = append(subscribers, <-newSubscribers)
        }
        close(topicIsReady)//alert listeners that go-routine is running
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
            case event:=<-events:
                if closed {
                    if logger != nil {
                        logger(fmt.Sprintf("%v closed.", name))
                    }
                    return
                }
                //note: when channel is closed event == nil
                if event != nil {
                    if logger != nil {
                        logger(fmt.Sprintf("%v notifying %v subscribers about %v.", name, len(subscribers), event))
                    }
                    for _, subscriber := range subscribers {
                        // log.Println(fmt.Sprintf("%v <- %v %T", t.String(), event, event))
                        //note: if subscriber sends something to a channel we don't want to be blocked.
                        go subscriber(event)
                    }
                }
            }
        }
    }()
    return topicIsReady
}

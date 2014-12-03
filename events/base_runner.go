package events

import (
)

/**
This function lies at the core of the base implementation of a Topic.

Essentially this is a state-machine, in which the provided channels change the state.
*/
func buildBaseLoop(spec *topicSpec) func() {
    return func() {
        closed := false
        //note: line below is to make sure that subscribing occurs before ANY event publishing.
        if len(spec.subscribers) == 0 {
            spec.subscribers = append(spec.subscribers, <-spec.newSubscribers)
        }
        for ;; {
            select {
            case <-spec.finish: //released when you close the channel
                closed = true
                optionallyLog(spec, "is closed")
                return
            case newSubscriber:=<-spec.newSubscribers:
                if closed {
                    optionallyLog(spec, "is closed")
                    return
                }
                //note: when channel is closed newSubscriber == nil
                if newSubscriber != nil {
                    spec.subscribers = append(spec.subscribers, newSubscriber)
                }
            case event:=<-spec.events:
                if closed {
                    optionallyLog(spec, "is closed")
                    return
                }
                //note: when channel is closed event == nil
                if event != nil {
                    optionallyLog(spec, "notifying subscribers")
                    for _, subscriber := range spec.subscribers {
                        //note: if subscriber sends something to a channel we don't want to be blocked.
                        go subscriber(event)
                    }
                }
            }
        }
    }
}

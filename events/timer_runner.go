package events

import (
    "time"
)

/**
This function lies at the core of the Timer implementation of a Topic.

Essentially this is a state-machine, in which the provided channels change the state.
*/
func buildTimerLoop(spec *topicSpec, timeout time.Duration) func() {
    return func() {
        closed := false
        timer := time.After(timeout)
        //note: line below is to make sure that subscribing occurs before ANY event publishing.
        if len(spec.subscribers) == 0 {
            spec.subscribers = append(spec.subscribers, <-spec.newSubscribers)
        }
        for ;; {
            select {
            case <-spec.finish: //released when you close the channel
                closed = true
                return
            case newSubscriber:=<-spec.newSubscribers:
                if closed {
                    optionallyLog(spec, "closed")
                    return
                }
                //note: when channel is closed newSubscriber == nil
                if newSubscriber != nil {
                    spec.subscribers = append(spec.subscribers, newSubscriber)
                }
            case current:=<-timer:
                if closed {
                    optionallyLog(spec, "closed")
                    return
                }
                //note: when channel is closed event == nil
                optionallyLog(spec, "notifying subscribers")
                for _, subscriber := range spec.subscribers {
                    //note: if subscriber sends something to a channel we don't want to be blocked.
                    go subscriber(current)
                }
            case <-spec.events:
                if closed {
                    optionallyLog(spec, "closed")
                    return
                }
                timer = time.After(timeout)//resets timer
            }
        }
    }
}

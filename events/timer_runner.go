package events

import (
    "time"
)

/**
This function lies at the core of the Timer implementation of a Topic.

Essentially this is a state-machine, in which the provided channels change the state.
*/
func buildTickerLoop(spec *topicSpec, timeout time.Duration) func() {
    return func() {
        closed := false
        timer := time.NewTicker(timeout)
        //note: line below is to make sure that subscribing occurs before ANY event publishing.
        if len(spec.subscribers) == 0 {
            spec.subscribers = append(spec.subscribers, <-spec.newSubscribers)
            optionallyLog(spec, "added subscriber")
        }
        optionallyLog(spec, "has started")
        for ;; {
            select {
            case <-spec.finish: //released when you close the channel
                timer.Stop()
                closed = true
                optionallyLog(spec, "is closed")
                return
            case newSubscriber:=<-spec.newSubscribers:
                if closed {
                    optionallyLog(spec, "has just closed, so subscriber can't be added")
                    return
                }
                //note: when channel is closed newSubscriber == nil
                if newSubscriber != nil {
                    spec.subscribers = append(spec.subscribers, newSubscriber)
                }
                optionallyLog(spec, "added subscriber")
            case current:=<-timer.C:
                if closed {
                    optionallyLog(spec, "has just closed, so ticking is abandoned")
                    return
                }
                //note: when channel is closed event == nil
                for _, subscriber := range spec.subscribers {
                    //note: if subscriber sends something to a channel we don't want to be blocked.
                    go subscriber(current)
                }
                optionallyLog(spec, "notified subscribers")
            case <-spec.events:
                if closed {
                    optionallyLog(spec, "has closed, so abandoning the nudge")
                    return
                }
                timer.Stop()
                timer = time.NewTicker(timeout)//resets timer
            }
        }
    }
}

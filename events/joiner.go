package events

import (
    "time"
    "fmt"
)

/* A shorthand for a map of events, each parameterised by a key */
type NamedEvents map[string]interface{}

/* Allows you to subscribe to multiple Topics at once, and wait until all of them have been notified by a Publish. Do note, that due to the architecture, the function may wait indefinitely, if one of the Topics does not have a Publish.*/
func AwaitAll(pendingEvents []Topic, waitFor time.Duration) <-chan NamedEvents {
    return await(pendingEvents, waitFor, false)
}

/* Allows you to subscribe to multiple Topics at once, and wait until all of them have been notified by a Publish. Do note, the function may panic after the specified Duration if it events have not appeared in all provided Topics. */
func MustAwaitAll(pendingEvents []Topic, waitFor time.Duration) <-chan NamedEvents {
    return await(pendingEvents, waitFor, true)
}

func await(pendingEvents []Topic, waitFor time.Duration, panicOnTimeout bool) <-chan NamedEvents {
    events := map[string]interface{} {}
    pending := make(chan NamedEvents)
    releaser := make(chan NamedEvents)
    newSubscriber := func(name string) func(event interface{}) {
        return func(event interface{}) {
            pending <- map[string]interface{} { name : event }
        }
    }
    for _, topic := range pendingEvents {
        if topic == nil {
            panic(fmt.Sprintf("No topic provided. Array has empty fields: %v ", pendingEvents))
        }
        topic.NewSubscriber(newSubscriber(topic.String()))
    }
    go func() {
        for ;; {
            select {
            case namedEvent := <-pending:
                for key,value := range namedEvent {
                    events[key] = value
                }
                if len(events) == len(pendingEvents) {
                    releaser<-events
                    return
                }
            case <-time.After(waitFor):
                if (panicOnTimeout) {
                    panic("Some expected events have not been published in the expected timeframe")
                }
                return
            }
        }
    }()
    return releaser
}


package joiner

import (
    "events"
    "log"
)

type NamedEvents map[string]interface{}

func AwaitAll(pendingEvents []events.Topic) <-chan NamedEvents {
    events := map[string]interface{} {}
    pending := make(chan NamedEvents)
    releaser := make(chan NamedEvents)
    newSubscriber := func(name string) func(event interface{}) {
        return func(event interface{}) {
            pending <- map[string]interface{} { name : event }
        }
    }
    for _, topic := range pendingEvents {
        topic.NewSubscriber(newSubscriber(topic.String()))
    }
    go func() {
        for ;; {
            namedEvent := <-pending
            for key,value := range namedEvent {
                events[key] = value
            }
            log.Println(events)
            log.Println(len(pendingEvents))
            if len(events) == len(pendingEvents) {
                releaser<-events
                return
            }
        }
    }()
    return releaser
}



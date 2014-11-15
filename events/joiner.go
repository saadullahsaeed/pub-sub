package events

import (
    "time"
    "fmt"
    "errors"
)

/* Represents results of calling AwaitAll, MustAwaitAll, AwaitOr, MustAwairOr. Technically, it is a container for a map of events (each parameterised by a Topic key) and an optional error */
type Result struct {
    events map[string]interface{}
    err error
}

/* Allows you to subscribe to multiple Topics at once, and wait until ALL of them have been notified by a Publish. Do note, that due to the architecture, the function may wait indefinitely, if one of the Topics does not have a Publish.*/
func AwaitAll(pendingEvents []Topic, waitFor time.Duration) <-chan *Result {
    return await(pendingEvents, waitFor, verifyIfAllEventsOccured)
}

/* Allows you to subscribe to multiple Topics at once, and wait until ANY of them have been notified by a Publish. Do note, that due to the architecture, the function may wait indefinitely, if one of the Topics does not have a Publish.*/
func AwaitOr(pendingEvents []Topic, waitFor time.Duration) <-chan *Result {
    return await(pendingEvents, waitFor, verifyIfAnyEventOccured)
}

/* Allows you to subscribe to multiple Topics at once, and wait until ALL of them have been notified by a Publish. Do note, the function may panic after the specified Duration if it events have not appeared in all provided Topics. */
func MustAwaitAll(pendingEvents []Topic, waitFor time.Duration) map[string]interface{} {
    result := <-await(pendingEvents, waitFor, verifyIfAllEventsOccured)
    if result.err != nil {
        panic(result.err.Error())
    }
    return result.events
}

/* Allows you to subscribe to multiple Topics at once, and wait until ANY of them have been notified by a Publish. Do note, the function may panic after the specified Duration if it events have not appeared in all provided Topics. */
func MustAwaitOr(pendingEvents []Topic, waitFor time.Duration) map[string]interface{} {
    result := <-await(pendingEvents, waitFor, verifyIfAnyEventOccured)
    if result.err != nil {
        panic(result.err.Error())
    }
    return result.events
}

func await(pendingEvents []Topic, waitFor time.Duration, verifier func(map[string]interface{}, []Topic) bool) <-chan *Result {
    events := map[string]interface{} {}
    pending := make(chan map[string]interface{})
    releaser := make(chan *Result)
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
                if verifier(events, pendingEvents) {
                    releaser<- &Result { events, nil }
                    return
                }
            case <-time.After(waitFor):
                notFound := []string {}
                for _, value := range pendingEvents {
                    if _, exists := events[value.String()]; !exists {
                        notFound = append(notFound, value.String())
                    }
                }
                releaser<- &Result { events, errors.New(fmt.Sprintf("Some events have not been published in the expected timeframe: %v", notFound)) }
                return
            }
        }
    }()
    return releaser
}

func verifyIfAllEventsOccured(events map[string]interface{}, pendingEvents []Topic) bool {
    return len(events) == len(pendingEvents)
}

func verifyIfAnyEventOccured(events map[string]interface{}, pendingEvents []Topic) bool {
    return len(events) > 0
}


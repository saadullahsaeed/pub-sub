package events

import (
    "time"
)

type spec struct {
    name string
    loggingMethod func(string, ...interface{})
}

type timeoutSpec struct {
    name string
    timeout time.Duration
}

type subscriberSpec struct {
    name string
    subscriber Subscriber
}

type eventSpec struct {
    name string
    event interface{}
	delay time.Duration //if the value is negative, there is no requeue
}

type stateModifierSpec struct {
    modifier func(snapshot *factory)
    stateChanged chan bool
    kill bool
}

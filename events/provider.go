package events
import (
    "time"
    "fmt"
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
}

type provider struct {
    newSpecs chan *spec
    newTimeouts chan *timeoutSpec
    newSubscribers chan *subscriberSpec
    topics map[string]Topic
    subscribers map[string][]Subscriber
    events chan *eventSpec
    closeEvents chan string
}

func (t *provider) NewTopic(topicName string) Topic {
    return &providerTopic {
        t,
        topicName,
    }
}

func (t *provider) NewTopicWithLogging(topicName string, loggingMethod func(string, ...interface{})) Topic {
    return &providerTopic {
        t,
        topicName,
    }
}

func (t *provider) Close() error {
    return nil
}

func (t *provider) String() string {
    return fmt.Sprintf("Topic-provider {size=%v}", len(t.topics))
}

type providerTopic struct {
    p *provider
    name string
}

func (t *providerTopic) String() string {
    return t.name
}

func (t *providerTopic) NewPublisher() Publisher {
    publisher := func(event interface{}) {
        //it's crucial this is in a go-routine: running 2+ Publishers in the same
        //go-routine causes a deadlock without this.
        go func() {
            t.p.events<- &eventSpec { t.name, event }
        }()
    }
    return publisher
}

func (t *providerTopic) NewSubscriber(subscriber Subscriber) <-chan bool {
    releaser := make(chan bool)
    go func() {
        t.p.newSubscribers<-&subscriberSpec { t.name, subscriber }
        close(releaser) //this releases awaiting listeners
    }()
    return releaser
}

func (t *providerTopic) Close() error {
    go func() {
        t.p.closeEvents<-t.String()
    }()
    return nil
}
func NewProvider() Provider {
    topicProvider := &provider {
        make(chan *spec),
        make(chan *timeoutSpec),
        make(chan *subscriberSpec),
        map[string]Topic {},
        map[string][]Subscriber {},
        make(chan *eventSpec),
        make(chan string),
    }
    <-runProvider(topicProvider)
    return topicProvider
}

func runProvider(p *provider) <-chan bool {
    releaser := make(chan bool)
    go func() {
        close(releaser)
        for ;; {
            select {
            case name:= <-p.closeEvents:
                delete(p.topics, name)
                delete(p.subscribers, name)
            case newSubscriber:=<-p.newSubscribers:
                if newSubscriber != nil {
                    p.subscribers[newSubscriber.name] = append(p.subscribers[newSubscriber.name],newSubscriber.subscriber)
                }
            case event:=<-p.events:
                if subscribers, subscribersExist := p.subscribers[event.name]; subscribersExist {
                    for _, subscriber := range subscribers {
                        //note: if subscriber sends something to a channel we don't want to be blocked.
                        go subscriber(event.event)
                    }
                } else {
                    go reQueue(event, p.events)
                }

            }
        }
    }()
    return releaser
}

func reQueue(e *eventSpec, ch chan *eventSpec) {
    ch<-e
}

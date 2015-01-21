package events
import (
    "time"
    "fmt"
    "log"
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

type stateModifierSpec struct {
    modifier func(snapshot *provider)
}

type simpleTopic struct {
    p *provider
    name string
    optionalState interface{}
}

func (t *simpleTopic) String() string {
    return fmt.Sprintf("%v { %v }", t.name, t.optionalState)
}

func (t *simpleTopic) NewPublisher() Publisher {
    publisher := func(event interface{}) {
        //it's crucial this is in a go-routine: running 2+ Publishers in the same
        //go-routine causes a deadlock without this.
        go func() {
            log.Println(fmt.Sprintf("%v <- %v", t.name, event))
            t.p.events<- &eventSpec { t.name, event }
        }()
    }
    return publisher
}

func (t *simpleTopic) NewSubscriber(subscriber Subscriber) <-chan bool {
    releaser := make(chan bool)
    go func() {
        t.p.newSubscribers<-&subscriberSpec { t.name, subscriber }
        close(releaser) //this releases awaiting listeners
    }()
    return releaser
}

func (t *simpleTopic) Close() error {
    remover := func(state *provider) {
        delete(state.topics, t.name)
        delete(state.subscribers, t.name)
    }
    t.p.stateModifier <- &stateModifierSpec { remover }
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
        make(chan *stateModifierSpec),
    }
    <-runProvider(topicProvider)
    return topicProvider
}

type provider struct {
    newSpecs chan *spec
    newTimeouts chan *timeoutSpec
    newSubscribers chan *subscriberSpec
    topics map[string]Topic
    subscribers map[string][]Subscriber
    events chan *eventSpec
    stateModifier chan *stateModifierSpec
}

func (t *provider) NewTopic(topicName string) Topic {
    topic := &simpleTopic { t, topicName, nil }
    adder := func(state *provider) {
        state.topics[topicName] = topic
        state.subscribers[topicName] = []Subscriber {}
    }
    t.stateModifier <- &stateModifierSpec { adder }
    return topic
}

func (t *provider) NewTopicWithLogging(topicName string, loggingMethod func(string, ...interface{})) Topic {
    topic := &simpleTopic { t, topicName, nil }
    adder := func(state *provider) {
        state.topics[topicName] = topic
        state.subscribers[topicName] = []Subscriber {}
    }
    t.stateModifier <- &stateModifierSpec { adder }
    return topic
}

func (t *provider) buildAndSubscriber(andTopic *simpleTopic, topic Topic, topics []Topic) Subscriber {
    return func(event interface{}) {
        stateModifier := func(pt *provider) {
            results := andTopic.optionalState.(map[string][]interface{})
            results[topic.String()] = append(results[topic.String()], event)
            if len(results) == len(topics) {
                andTopic.NewPublisher()(copyAside(results))
                // andTopic.optionalState = map[string][]interface{} {}
            } else {
                andTopic.optionalState = results
            }
        }
        t.stateModifier <- &stateModifierSpec { stateModifier }
    }
}

func (t *provider) JoinWithAnd(topics []Topic, name string) Topic {
    releaser := make(chan Topic)
    adder := func(p *provider) {
        topicName := fmt.Sprintf("%v", topics)
        newTopic := &simpleTopic { t, topicName, map[string][]interface{} {} }
        p.topics[topicName] = newTopic
        p.subscribers[topicName] = []Subscriber {}
        for _, topic := range topics {
            topic.NewSubscriber(t.buildAndSubscriber(newTopic, topic, topics))
        }
        releaser<-newTopic
    }
    t.stateModifier <- &stateModifierSpec { adder }
    topic := <-releaser
    close(releaser)
    return topic
}

func (t *provider) Close() error {
    return nil
}

func (t *provider) String() string {
    return fmt.Sprintf("Topic-provider {size=%v}", len(t.topics))
}

func runProvider(p *provider) <-chan bool {
    releaser := make(chan bool)
    go func() {
        close(releaser)
        for ;; {
            select {
            case stateChange := <-p.stateModifier:
                stateChange.modifier(p)
                log.Println(fmt.Sprintf("state change: %v", p.topics))
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

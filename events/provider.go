package events
import (
    "time"
    "fmt"
    "log"
)

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

func (t *provider) NewTickerTopic(topicName string, interval time.Duration) Topic {
    topic := &tickerTopic { t, topicName, time.NewTicker(interval), make(chan bool) }
    adder := func(state *provider) {
        state.topics[topicName] = topic
        state.subscribers[topicName] = []Subscriber {}
        <-runTicker(topic, t)
    }
    t.stateModifier <- &stateModifierSpec { adder }
    return topic
}

func (t *provider) buildAndGateSubscriber(andTopic *simpleTopic, topic Topic, topics []Topic) Subscriber {
    return func(event interface{}) {
        stateModifier := func(pt *provider) {
            results := andTopic.optionalState.(map[string][]interface{})
            results[topic.String()] = append(results[topic.String()], event)
            if len(results) == len(topics) {
                andTopic.NewPublisher()(copyAside(results))
                andTopic.optionalState = map[string][]interface{} {}
            } else {
                andTopic.optionalState = results
            }
        }
        t.stateModifier <- &stateModifierSpec { stateModifier }
    }
}

func (t *provider) buildOrGateSubscriber(orTopic *simpleTopic, topic Topic, topics []Topic) Subscriber {
    return func(event interface{}) {
        stateModifier := func(pt *provider) {
            results := orTopic.optionalState.(map[string][]interface{})
            results[topic.String()] = append(results[topic.String()], event)
            orTopic.NewPublisher()(copyAside(results))
            orTopic.optionalState = map[string][]interface{} {}
        }
        t.stateModifier <- &stateModifierSpec { stateModifier }
    }
}

func (t *provider) buildGateTopic(topics []Topic, subscriberFactory func(*simpleTopic, Topic, []Topic) Subscriber) Topic {
    releaser := make(chan Topic)
    adder := func(p *provider) {
        topicName := fmt.Sprintf("%v", topics)
        newTopic := &simpleTopic { t, topicName, map[string][]interface{} {} }
        p.topics[topicName] = newTopic
        p.subscribers[topicName] = []Subscriber {}
        for _, topic := range topics {
            topic.NewSubscriber(subscriberFactory(newTopic, topic, topics))
        }
        releaser<-newTopic
    }
    t.stateModifier <- &stateModifierSpec { adder }
    topic := <-releaser
    close(releaser)
    return topic
}

func (t *provider) OrGate(topics []Topic) Topic {
    return t.buildGateTopic(topics, t.buildOrGateSubscriber)
}

func (t *provider) AndGate(topics []Topic) Topic {
    return t.buildGateTopic(topics, t.buildAndGateSubscriber)
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

func copyAside(original map[string][]interface{}) map[string][]interface{} {
    mapCopy := map[string][]interface{} {}
    for key,value := range original {
        copiedValue := []interface{} {}
        for _, arrayValue := range value {
            copiedValue = append(copiedValue, arrayValue)
        }
        mapCopy[key] = copiedValue
    }
    return mapCopy
}

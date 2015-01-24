package events
import (
    "time"
    "fmt"
    // "log"
)

func NewFactory() Factory {
    topicFactory := &factory {
        make(chan *spec),
        make(chan *timeoutSpec),
        make(chan *subscriberSpec),
        map[string]Topic {},
        map[string][]Subscriber {},
        make(chan *eventSpec),
        make(chan *stateModifierSpec),
    }
    <-runFactory(topicFactory)
    return topicFactory
}

type factory struct {
    newSpecs chan *spec
    newTimeouts chan *timeoutSpec
    newSubscribers chan *subscriberSpec
    topics map[string]Topic
    subscribers map[string][]Subscriber
    events chan *eventSpec
    stateModifier chan *stateModifierSpec
}

func (t *factory) NewTopic(topicName string) Topic {
    topic := &simpleTopic { t, topicName, nil }
    adder := func(state *factory) {
        state.topics[topicName] = topic
        state.subscribers[topicName] = []Subscriber {}
    }
    t.stateModifier <- &stateModifierSpec { adder }
    return topic
}

func (t *factory) NewTickerTopic(topicName string, interval time.Duration) Topic {
    topic := &tickerTopic { t, topicName, time.NewTicker(interval), make(chan bool) }
    adder := func(state *factory) {
        state.topics[topicName] = topic
        state.subscribers[topicName] = []Subscriber {}
        <-runTicker(topic, t)
    }
    t.stateModifier <- &stateModifierSpec { adder }
    return topic
}

func (t *factory) buildAndGateSubscriber(andTopic *simpleTopic, topic Topic, topics []Topic) Subscriber {
    return func(event interface{}) {
        stateModifier := func(pt *factory) {
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

func (t *factory) buildOrGateSubscriber(orTopic *simpleTopic, topic Topic, topics []Topic) Subscriber {
    return func(event interface{}) {
        stateModifier := func(pt *factory) {
            results := orTopic.optionalState.(map[string][]interface{})
            results[topic.String()] = append(results[topic.String()], event)
            orTopic.NewPublisher()(copyAside(results))
            orTopic.optionalState = map[string][]interface{} {}
        }
        t.stateModifier <- &stateModifierSpec { stateModifier }
    }
}

func (t *factory) buildGateTopic(topics []Topic, subscriberFactory func(*simpleTopic, Topic, []Topic) Subscriber, separator string) Topic {
    releaser := make(chan Topic)
    adder := func(p *factory) {
        topicName := ""
        for _, topic := range topics {
            topicName = topicName + separator + topic.String()
        }
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

func (t *factory) OrGate(topics []Topic) Topic {
    return t.buildGateTopic(topics, t.buildOrGateSubscriber, "|")
}

func (t *factory) AndGate(topics []Topic) Topic {
    return t.buildGateTopic(topics, t.buildAndGateSubscriber,"&")
}

func (t *factory) Close() error {
    return nil
}

func (t *factory) String() string {
    return fmt.Sprintf("Topic-factory {size=%v}", len(t.topics))
}

func runFactory(p *factory) <-chan bool {
    releaser := make(chan bool)
    go func() {
        close(releaser)
        for ;; {
            select {
            case stateChange := <-p.stateModifier:
                stateChange.modifier(p)
                // log.Println(fmt.Sprintf("state change: %v", p.topics))
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
    //this value has been fined tuned by running some benchmarks.
    //it seems that with a lower value, the publish/subscribe events are being offset bh the subscriber not being there yet. 
    <-time.After(time.Duration(5000))
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

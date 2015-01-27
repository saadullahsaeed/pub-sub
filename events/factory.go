package events
import (
    "time"
    "fmt"
    // "log"
)

func NewFactory() Factory {
    timeout := 5000
    //this value has been fined tuned by running some benchmarks.
    //it is most probably CPU/system dependent. There is a sweet-spot, though. 
    //on the laptops I use I got the basic benchmarks down to 2000ns/op, but only after raising the value up. 
    //it is clear though, that with a much bigger value than that above, the benchmark results will reverse and worsen.
    topicFactory := &factory {
        map[string]Topic {},
        map[string][]Subscriber {},
        make(chan *eventSpec),
        make(chan *stateModifierSpec),
        time.After(time.Duration(timeout)),
    }
    <-runFactory(topicFactory)
    return topicFactory
}

type factory struct {
    topics map[string]Topic
    subscribers map[string][]Subscriber
    events chan *eventSpec
    stateModifier chan *stateModifierSpec
    internalClock <-chan time.Time
}

func (t *factory) NewTopic(topicName string) Topic {
    topic := &simpleTopic { t, topicName, nil }
    stateChanged := make(chan bool)
    adder := func(state *factory) {
        state.topics[topicName] = topic
        state.subscribers[topicName] = []Subscriber {}
    }
    t.stateModifier <- &stateModifierSpec { adder, stateChanged, false }
    <-stateChanged
    close(stateChanged)
    return topic
}

func (t *factory) NewTickerTopic(topicName string, interval time.Duration) Topic {
    topic := &tickerTopic { t, topicName, time.NewTicker(interval), make(chan bool) }
    stateChanged := make(chan bool)
    adder := func(state *factory) {
        state.topics[topicName] = topic
        state.subscribers[topicName] = []Subscriber {}
        <-runTicker(topic, t)
    }
    t.stateModifier <- &stateModifierSpec { adder, stateChanged, false }
    <-stateChanged
    close(stateChanged)
    return topic
}

func (t *factory) buildAndGateSubscriber(andTopic *simpleTopic, topic Topic, topics []Topic) Subscriber {
    return func(event interface{}) {
        stateChanged := make(chan bool)
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
        t.stateModifier <- &stateModifierSpec { stateModifier, stateChanged, false }
        <-stateChanged
        close(stateChanged)
    }
}

func (t *factory) buildOrGateSubscriber(orTopic *simpleTopic, topic Topic, topics []Topic) Subscriber {
    return func(event interface{}) {
        stateChanged := make(chan bool)
        stateModifier := func(pt *factory) {
            results := orTopic.optionalState.(map[string][]interface{})
            results[topic.String()] = append(results[topic.String()], event)
            orTopic.NewPublisher()(copyAside(results))
            orTopic.optionalState = map[string][]interface{} {}
        }
        t.stateModifier <- &stateModifierSpec { stateModifier, stateChanged, false }
        <-stateChanged
        close(stateChanged)
    }
}

func (t *factory) buildGateTopic(topics []Topic, subscriberFactory func(*simpleTopic, Topic, []Topic) Subscriber, separator string) Topic {
    var (
        newTopic *simpleTopic
    )
    stateChanged := make(chan bool)
    adder := func(p *factory) {
        topicName := ""
        for _, topic := range topics {
            topicName = topicName + separator + topic.String()
        }
        newTopic = &simpleTopic { t, topicName, map[string][]interface{} {} }
        p.topics[topicName] = newTopic
        p.subscribers[topicName] = []Subscriber {}
        for _, topic := range topics {
            go topic.NewSubscriber(subscriberFactory(newTopic, topic, topics))
        }
    }
    t.stateModifier <- &stateModifierSpec { adder, stateChanged, false }
    <-stateChanged
    close(stateChanged)
    return newTopic
}

func (t *factory) OrGate(topics []Topic) Topic {
    return t.buildGateTopic(topics, t.buildOrGateSubscriber, "|")
}

func (t *factory) AndGate(topics []Topic) Topic {
    return t.buildGateTopic(topics, t.buildAndGateSubscriber,"&")
}

func (t *factory) Close() error {
    //we close topics manually here, otherwise you may get a deadlock 
    stateChanged := make(chan bool)
    closer := func(p *factory) {
        for _, topic := range t.topics {
            delete(t.topics, topic.String())
            delete(t.subscribers, topic.String())
        }
    }
    t.stateModifier <- &stateModifierSpec { closer, stateChanged, true }
    <-stateChanged
    close(stateChanged)
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
                stateChange.stateChanged<-true
                if stateChange.kill {
                    break
                }
            case event:=<-p.events:
                if subscribers, subscribersExist := p.subscribers[event.name]; subscribersExist {
                    for _, subscriber := range subscribers {
                        //note: if subscriber sends something to a channel we don't want to be blocked.
                        go subscriber(event.event)
                    }
                } else {
                    go p.reQueue(event)
                }
            }
        }
        close(p.events)
        close(p.stateModifier)
    }()
    return releaser
}

func (t *factory) reQueue(e *eventSpec) {
    <-t.internalClock
    t.events<-e
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

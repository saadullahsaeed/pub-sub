package events

import (
    "time"
)
type tickerTopic struct {
    p *factory
    name string
    ticker *time.Ticker
    closeChannel chan bool
}

func (t *tickerTopic) String() string {
    return t.name
}

func (t *tickerTopic) NewPublisher() Publisher {
    panic("Tickers can't be published to")
}

func (t *tickerTopic) NewSubscriber(subscriber Subscriber) {
    stateChanged := make(chan bool)
    adder := func(p *factory) {
        if subscriber != nil {
            p.subscribers[t.name] = append(p.subscribers[t.name], subscriber)
        }
    }
    t.p.stateModifier <- &stateModifierSpec { adder, stateChanged }
    <-stateChanged
    close(stateChanged)
}

func (t *tickerTopic) Close() error {
    stateChanged := make(chan bool)
    remover := func(state *factory) {
        delete(state.topics, t.name)
        delete(state.subscribers, t.name)
        close(t.closeChannel)
        t.ticker.Stop()
    }
    t.p.stateModifier <- &stateModifierSpec { remover, stateChanged }
    <-stateChanged
    close(stateChanged)
    return nil
}

func runTicker(topic *tickerTopic, t *factory) <-chan bool {
    releaser := make(chan bool)
    go func() {
        close(releaser)
        for ;; {
            select {
            case <-topic.closeChannel:
                return
            case snapshot := <-topic.ticker.C:
                go func() {
                    t.events<- &eventSpec { topic.name, snapshot }
                }()
            }
        }
    }()
    return releaser
}


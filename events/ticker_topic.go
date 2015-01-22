package events

import (
    "time"
)
type tickerTopic struct {
    p *provider
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

func (t *tickerTopic) NewSubscriber(subscriber Subscriber) <-chan bool {
    releaser := make(chan bool)
    go func() {
        t.p.newSubscribers<-&subscriberSpec { t.name, subscriber }
        close(releaser) //this releases awaiting listeners
    }()
    return releaser
}

func (t *tickerTopic) Close() error {
    remover := func(state *provider) {
        delete(state.topics, t.name)
        delete(state.subscribers, t.name)
        close(t.closeChannel)
        t.ticker.Stop()
    }
    t.p.stateModifier <- &stateModifierSpec { remover }
    return nil
}

func runTicker(topic *tickerTopic, t *provider) <-chan bool {
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


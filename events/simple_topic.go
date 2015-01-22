package events

type simpleTopic struct {
    p *factory
    name string
    optionalState interface{}
}

func (t *simpleTopic) String() string {
    return t.name
}

func (t *simpleTopic) NewPublisher() Publisher {
    publisher := func(event interface{}) {
        //it's crucial this is in a go-routine: running 2+ Publishers in the same
        //go-routine causes a deadlock without this.
        go func() {
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
    remover := func(state *factory) {
        delete(state.topics, t.name)
        delete(state.subscribers, t.name)
    }
    t.p.stateModifier <- &stateModifierSpec { remover }
    return nil
}


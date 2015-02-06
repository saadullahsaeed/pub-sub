package events

type simpleTopic struct {
	p             *factory
	name          string
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
			t.p.events <- &eventSpec{t.name, event}
		}()
	}
	return publisher
}

func (t *simpleTopic) NewSubscriber(subscriber Subscriber) {
	stateChanged := make(chan bool)
	adder := func(p *factory) {
		if subscriber != nil {
			p.subscribers[t.name] = append(p.subscribers[t.name], subscriber)
		}
	}
	t.p.stateModifier <- &stateModifierSpec{adder, stateChanged, false}
	<-stateChanged
	close(stateChanged)
}

func (t *simpleTopic) Close() error {
	stateChanged := make(chan bool)
	remover := func(state *factory) {
		if _, exists := state.topics[t.name]; exists {
			delete(state.topics, t.name)
			delete(state.subscribers, t.name)
		}
	}
	t.p.stateModifier <- &stateModifierSpec{remover, stateChanged, false}
	<-stateChanged
	close(stateChanged)
	return nil
}

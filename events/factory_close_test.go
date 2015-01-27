package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "time"
)

func TestThat_YouCanCloseFactory(t *testing.T) {
    assert := assertions.New(t)
    factory := NewFactory()
    factory.NewTopic("hello")

    assert.DoesNotThrow(func() {
        factory.Close()
    })
}

func TestThat_YouCanCloseFactory_And_Topics(t *testing.T) {
    assert := assertions.New(t)
    factory := NewFactory()
    topic := factory.NewTopic("hello")

    topic.Close()

    assert.DoesNotThrow(func() {
        factory.Close()
    })
}

func TestThat_AfterClosing_TopicsCantBeUsed(t *testing.T) {
    assert := assertions.New(t)
    factory := NewFactory()
    topic := factory.NewTopic("hello")
    topic.NewSubscriber(func(event interface{}) {
        //should not execute really
        assert.IsTrue(false)
    })

    assert.DoesNotThrow(func() {
        factory.Close()
    })
    topic.NewPublisher()("hello")
    //allow the go-routine to pick up things
    <-time.After(time.Duration(10)*time.Millisecond)
}

package handlers

import (
    "testing"
    "github.com/tholowka/testing/assertions"
)

func TestThat_PubSub_Works(t *testing.T) {
    //given
    assert := assertions.New(t)
    topic := NewTopic("my-awesome-rant")
    channel := make(chan string)
    publisher := topic.NewPublisher()
    subscriber := func(event interface{}) {
        channel<-event.(string)
    }
    topic.NewSubscriber(subscriber)
    //when
    go publisher("and I get to talk about jazz")
    //then the subscriber actually got invoked since the channel received some news
    assert.AreEqual("and I get to talk about jazz", <-channel)
}

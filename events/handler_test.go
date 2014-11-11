package events

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

func TestThat_MultiplePublishers_Work(t *testing.T) {
    //given
    assert := assertions.New(t)
    topic := NewTopic("my-awesome-rant")
    channel := make(chan string)
    firstPublisher := topic.NewPublisher()
    secondPublisher := topic.NewPublisher()
    subscriber := func(event interface{}) {
        channel<-event.(string)
    }
    topic.NewSubscriber(subscriber)
    //when
    go firstPublisher("and I get to talk about jazz")
    go secondPublisher("and I get to talk about bebop")
    //then 
    assert.AreEqual("and I get to talk about jazz", <-channel)
    assert.AreEqual("and I get to talk about bebop", <-channel)
}

func TestThat_MultipleSubscribers_Work(t *testing.T) {
    //given
    assert := assertions.New(t)
    topic := NewTopic("my-awesome-rant")
    channel := make(chan string)
    publisher := topic.NewPublisher()
    firstSubscriber := func(event interface{}) {
        channel<-"one"
    }
    secondSubscriber := func(event interface{}) {
        channel<-"two"
    }
    topic.NewSubscriber(firstSubscriber)
    topic.NewSubscriber(secondSubscriber)
    //when
    go publisher("was Charlie better than John")
    //then the subscriber actually got invoked since the channel received some news
    firstResult := <-channel
    secondResult := <-channel
    assert.AreEqual([]string { firstResult, secondResult}, []string{ "one", "two"})
}

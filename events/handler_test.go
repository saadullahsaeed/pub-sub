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
    publisher := topic.NewPublisher(nil)
    subscriber := func(event interface{}) {
        channel<-event.(string)
    }
    topic.NewSubscriber(subscriber)
    //when
    publisher("and I get to talk about jazz")
    //then the subscriber actually got invoked since the channel received some news
    assert.AreEqual("and I get to talk about jazz", <-channel)
    topic.Close()
}

func TestThat_MultiplePublishers_Work(t *testing.T) {
    //given
    assert := assertions.New(t)
    topic := NewTopic("my-awesome-rant")
    channel := make(chan string)
    firstPublisher := topic.NewPublisher(nil)
    secondPublisher := topic.NewPublisher(nil)
    subscriber := func(event interface{}) {
        channel<-event.(string)
    }
    topic.NewSubscriber(subscriber)
    //when
    firstPublisher("and I get to talk about jazz")
    secondPublisher("and I get to talk about bebop")
    //then 
    assert.AreEqual("and I get to talk about jazz", <-channel)
    assert.AreEqual("and I get to talk about bebop", <-channel)
    topic.Close()
}

func TestThat_MultipleSubscribers_Work(t *testing.T) {
    //given
    assert := assertions.New(t)
    topic := NewTopic("my-awesome-rant")
    channel := make(chan string)
    publisher := topic.NewPublisher(nil)
    firstSubscriber := func(event interface{}) {
        channel<-"one"
    }
    secondSubscriber := func(event interface{}) {
        channel<-"two"
    }
    topic.NewSubscriber(firstSubscriber)
    topic.NewSubscriber(secondSubscriber)
    //when
    publisher("was Charlie better than John")
    //then the subscriber actually got invoked since the channel received some news
    firstResult := <-channel
    secondResult := <-channel
    assert.AreEqual([]string { firstResult, secondResult}, []string{ "one", "two"})
    topic.Close()
}

func Benchmark_Propagation(b *testing.B) {
    topic := NewTopic("my-awesome-rant")
    subscriber := func(interface{}) {}
    topic.NewSubscriber(subscriber)
    for n:=0; n<b.N;n++ {
        topic.NewPublisher(nil)("Or is Keith J the best")
    }
    //note: closing the channel either skews test results or crashes a go-routine due to premature invocation.
}

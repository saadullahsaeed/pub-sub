package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
)

func TestThat_PubSub_Works(t *testing.T) {
    //given
    assert := assertions.New(t)
    topic := NewFactory().NewTopic("my-awesome-rant")
    channel := make(chan string)
    publisher := topic.NewPublisher()
    subscriber := func(event interface{}) {
        channel<-event.(string)
    }
    <-topic.NewSubscriber(subscriber)
    //when
    publisher("and I get to talk about jazz")
    //then the subscriber actually got invoked since the channel received some news
    assert.AreEqual("and I get to talk about jazz", <-channel)
    topic.Close()
}

func TestThat_MultiplePublishers_Work(t *testing.T) {
    //given
    assert := assertions.New(t)
    topic := NewFactory().NewTopic("my-public-rant")
    channel := make(chan string)
    firstPublisher := topic.NewPublisher()
    secondPublisher := topic.NewPublisher()
    subscriber := func(event interface{}) {
        channel<-event.(string)
    }
    <-topic.NewSubscriber(subscriber)
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
    topic := NewFactory().NewTopic("my-secret-rant")
    channel := make(chan string)
    publisher := topic.NewPublisher()
    firstSubscriber := func(event interface{}) {
        channel<-"one"
    }
    secondSubscriber := func(event interface{}) {
        channel<-"two"
    }
    <-topic.NewSubscriber(firstSubscriber)
    <-topic.NewSubscriber(secondSubscriber)
    //when
    publisher("was Charlie better than John")
    //then the subscriber actually got invoked since the channel received some news
    firstResult := <-channel
    secondResult := <-channel
    assert.AreEqual([]string { firstResult, secondResult}, []string{ "one", "two"})
    topic.Close()
}

func Benchmark_Propagation_When_CreatingPublishers_OnEachRequest(b *testing.B) {
    topic := NewFactory().NewTopic("my-awesome-rant")
    subscriber := func(interface{}) {}
    <-topic.NewSubscriber(subscriber)
    b.ResetTimer()
    for n:=0; n<b.N;n++ {
        topic.NewPublisher()("Or is Keith J the best")
    }
    //note: closing the channel either skews test results or crashes a go-routine due to premature invocation.
}

func Benchmark_Propagation_When_ReusingAPublisher_OnEachRequest(b *testing.B) {
    topic := NewFactory().NewTopic("my-awesome-rant")
    subscriber := func(interface{}) {}
    publisher := topic.NewPublisher()
    <-topic.NewSubscriber(subscriber)
    b.ResetTimer()
    for n:=0; n<b.N;n++ {
        publisher("Or is Marcus M the best")
    }
    //note: closing the channel either skews test results or crashes a go-routine due to premature invocation.
}

func Benchmark_Parallel_Topics(b *testing.B) {
    factory := NewFactory()
    topic := factory.NewTopic("my-awesome-rant")
    subscriber := func(interface{}) {}
    <-topic.NewSubscriber(subscriber)
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            topic.NewPublisher()("Or is Marcus M the best")
        }
    })
}




package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "time"
)

func TestThat_TemporaryTopics_CanBeCreated(t *testing.T) {
    //given
    assert := assertions.New(t)
    //then
    assert.IsNotNil(NewTemporaryTopic(NewTopic("a-topic"), time.Duration(10)*time.Millisecond))
}

func TestThat_TemporaryTopics_AreClosed(t *testing.T) {
    //given
    assert := assertions.New(t)
    subscribeOccured := make(chan string)
    channelErrorOccured := make(chan bool)
    duration := time.Duration(100)*time.Millisecond
    topic := NewTemporaryTopic(NewTopic("a-temporary-topic"), duration)
    topic.NewSubscriber(func(interface{}) {
        subscribeOccured<-"boom"
    })
    callback := func(event interface{}) {
        switch event.(type) {
        case error:
            channelErrorOccured<-true
        default:
            channelErrorOccured<-false
        }
    }
    //when 
    <-time.After(duration)
    topic.NewPublisher(callback)("Force a publish")
    //then
    select {
    case <-time.After(duration):
        assert.IsTrue(false)//should not occur
    case result:=<-channelErrorOccured:
        assert.IsTrue(result)
    case <-subscribeOccured:
        assert.IsTrue(false)//should not occur
    }
}

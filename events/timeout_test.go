package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "time"
)

func TestThat_WhenTimeout_Works(t *testing.T) {
    //given
    assert := assertions.New(t)
    topic := NewTopic("a-game-to-play")
    publisher := topic.NewPublisher(nil)
    waitForAnswer := make(chan bool)
    //when
    errorTopic := WhenTimeout(topic, time.Duration(50)*time.Millisecond, "timeouts")
    errorTopic.NewSubscriber(func(err interface{}) {
        switch err.(type) {
        case error:
            waitForAnswer<-true
        default:
            waitForAnswer<-false
        }
    })
    go func() {
        <-time.After(time.Duration(100)*time.Millisecond)
        publisher("hello")
    }()
    //then
    assert.IsTrue(<-waitForAnswer)
    errorTopic.Close()
}

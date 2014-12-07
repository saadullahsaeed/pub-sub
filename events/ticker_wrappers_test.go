package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "time"
    "log"
)

func TestThat_WhenTimeout_Works(t *testing.T) {
    //given
    delay := 100
    assert := assertions.New(t)
    topic := NewTopic("a-game-to-play")
    waitForAnswer := make(chan bool)
    //when
    errorTopic := WhenTimeoutWithLogging(topic, time.Duration(delay)*time.Millisecond, "timeouts-are-working", log.Println)
    <-errorTopic.NewSubscriber(func(err interface{}) {
        switch err.(type) {
        case time.Time:
            waitForAnswer<-true
        default:
            waitForAnswer<-false
        }
    })
    //then
    assert.IsTrue(<-waitForAnswer)
    <-time.After(time.Duration(6*delay))
    log.Println("Closing")
    errorTopic.Close()
}

func TestThat_WhenTimeout_Resets_EachTimeAnEventHappens(t *testing.T) {
    //given
    delay := 100
    assert := assertions.New(t)

    topic := NewTopic("a-game-to-play")
    publisher := topic.NewPublisher()
    waitForAnswer := make(chan time.Time)
    //when
    errorTopic := WhenTimeoutWithLogging(topic, time.Duration(5*delay)*time.Millisecond, "timeouts-do-reset", log.Println)
    <-errorTopic.NewSubscriber(func(err interface{}) {
        switch err.(type) {
        case time.Time:
            waitForAnswer<-err.(time.Time)
        default:
        }
    })
    startTime := time.Now()
    go func() {
        <-time.After(time.Duration(2*delay)*time.Millisecond)
        publisher("hello")
    }()
    //then
    log.Println(startTime)
    timeEnd := <-waitForAnswer
    timeElapsed := timeEnd.Sub(startTime)
    assert.IsNotNil(timeElapsed).IsTrue(timeElapsed > time.Duration(5*delay))
    errorTopic.Close()
}

func TestThat_CheckIfPublishOccured_ReturnsTrue_IfPublishedOccured(t *testing.T) {
    //given
    delay := 100
    startTest := make(chan bool)
    assert := assertions.New(t)

    topic := NewTopic("headhunters")
    publisher := topic.NewPublisher()
    go func() {
        //note this needs to be in a go-routine as Topics block until a Subscriber is registered
        close(startTest)
        publisher("chameleon")
    }()
    <-startTest
    result := <-CheckIfPublishOccuredAtLeastOnceWithLogging(topic, time.Duration(delay)*time.Millisecond, log.Println)
    assert.IsTrue(result)
}

func TestThat_CheckIfPublishOccured_ReturnsFalse_IfPublishedDidNotOccur(t *testing.T) {
    //given
    delay := 100
    assert := assertions.New(t)

    topic := NewTopic("headhunters")
    result := <-CheckIfPublishOccuredAtLeastOnceWithLogging(topic, time.Duration(delay)*time.Millisecond, log.Println)
    assert.IsFalse(result)
}

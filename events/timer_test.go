package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "time"
    "log"
)

func TestThat_TimerTopic_Pings(t *testing.T) {
    numberOfPings := 0
    duration := 100
    assert := assertions.New(t)

    timer := NewTickerTopicWithLogging("ticker", time.Duration(duration)*time.Millisecond, log.Println)
    <-timer.NewSubscriber(func(interface{}) {
         numberOfPings = numberOfPings+1
    })
    <-time.After(time.Duration(10*duration)*time.Millisecond)
    // assert.IsTrue(timer != nil)
    assert.IsTrue(numberOfPings>8)
    timer.Close()
}

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
    timeElapsed := <-waitForAnswer
    log.Println(timeElapsed)
    log.Println("Closing")
    assert.IsNotNil(timeElapsed)
    errorTopic.Close()
}

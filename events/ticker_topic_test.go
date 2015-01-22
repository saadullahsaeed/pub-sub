package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "time"
)

func TestThat_TimerTopic_Pings(t *testing.T) {
    numberOfPings := 0
    duration := 100
    assert := assertions.New(t)

    timer := NewFactory().NewTickerTopic("ticker", time.Duration(duration)*time.Millisecond)
    <-timer.NewSubscriber(func(interface{}) {
         numberOfPings = numberOfPings+1
    })
    <-time.After(time.Duration(10*duration)*time.Millisecond)
    // assert.IsTrue(timer != nil)
    assert.IsTrue(numberOfPings>8)
    timer.Close()
}



package events

import (
	"github.com/tholowka/testing/assertions"
	"testing"
	"time"
)

func TestThat_TimerTopic_Pings(t *testing.T) {
	//given
	numberOfPings := 0
	duration := 100
	assert := assertions.New(t)

	timer := NewFactory().NewTickerTopic("ticker", time.Duration(duration)*time.Millisecond)
	timer.NewSubscriber(func(interface{}) {
		numberOfPings = numberOfPings + 1
	})
	//when
	<-time.After(time.Duration(10*duration) * time.Millisecond)
	//then
	assert.IsTrue(numberOfPings > 8)
	timer.Close()
}

func TestThat_CallingCloseTwice_DoesNotHurt(t *testing.T) {
	//given
	timer := NewFactory().NewTickerTopic("ticker", time.Duration(100))
	assert := assertions.New(t)
	//when
	timer.Close()
	//then
	assert.DoesNotThrow(func() {
		timer.Close()
	})
}

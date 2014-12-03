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
//TODO
//Are temporary Topics closed?


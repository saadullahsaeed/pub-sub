package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "time"
)

func TestThat_IfYouDontWait_CodeDoesNotBlock(t *testing.T) {
    //given
    assert := assertions.New(t)
    firstTopic := NewTopic("rants-about-jazz")
    secondTopic := NewTopic("rants-about-bebop")
    //when
    releaser := AwaitAll([]Topic { firstTopic, secondTopic }, time.Duration(50)*time.Millisecond)
    //then
    assert.IsNotNil(releaser)
}

func TestThat_AwaitAll_BasicallyWorks(t *testing.T) {
    //given
    assert := assertions.New(t)
    firstTopic := NewTopic("rants-about-fusion")
    secondTopic := NewTopic("rants-about-bebop")
    //when
    releaser := AwaitAll([]Topic { firstTopic, secondTopic }, time.Duration(50)*time.Millisecond)
    firstTopic.NewPublisher()("Louis Armstrong was great!")
    secondTopic.NewPublisher()("The best duo was Charlie Parker and Diz in the 40s")
    //then
    result := (<-releaser).Events
    assert.IsNotNil(result).AreEqual("Louis Armstrong was great!", result["rants-about-fusion"]).
        AreEqual("The best duo was Charlie Parker and Diz in the 40s", result["rants-about-bebop"])
}

func TestThat_AwaitAll_SendsAnError_IfItCantComplete(t *testing.T) {
    //given
    assert := assertions.New(t)
    firstTopic := NewTopic("rants-about-fusion")
    secondTopic := NewTopic("rants-about-bebop")
    //when
    releaser := AwaitAll([]Topic { firstTopic, secondTopic }, time.Duration(50)*time.Millisecond)
    firstTopic.NewPublisher()("Louis Armstrong was great!")
    //then
    result := (<-releaser).Err
    assert.IsNotNil(result)
}

func TestThat_AwaitAny_BasicallyWorks(t *testing.T) {
    //given
    assert := assertions.New(t)
    firstTopic := NewTopic("rants-about-fusion")
    secondTopic := NewTopic("rants-about-bebop")
    //when
    releaser := AwaitAny([]Topic { firstTopic, secondTopic }, time.Duration(50)*time.Millisecond)
    firstTopic.NewPublisher()("Louis Armstrong was great!")
    //then
    result := (<-releaser).Events
    assert.IsNotNil(result).AreEqual("Louis Armstrong was great!", result["rants-about-fusion"])
}

func TestThat_AwaitAny_SendsAnError_IfItCantComplete(t *testing.T) {
    //given
    assert := assertions.New(t)
    firstTopic := NewTopic("rants-about-fusion")
    secondTopic := NewTopic("rants-about-bebop")
    //when
    releaser := AwaitAny([]Topic { firstTopic, secondTopic }, time.Duration(50)*time.Millisecond)
    //then
    result := (<-releaser).Err
    assert.IsNotNil(result)
}

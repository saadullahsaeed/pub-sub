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
    releaser := AwaitAll([]Topic { firstTopic, secondTopic }, time.Duration(10)*time.Second)
    //then
    assert.IsNotNil(releaser)
}

func TestThat_Joiner_BasicallyWorks(t *testing.T) {
    //given
    assert := assertions.New(t)
    firstTopic := NewTopic("rants-about-fusion")
    secondTopic := NewTopic("rants-about-bebop")
    //when
    releaser := AwaitAll([]Topic { firstTopic, secondTopic }, time.Duration(10)*time.Second)
    go firstTopic.NewPublisher()("Louis Armstrong was great!")
    go secondTopic.NewPublisher()("The best duo was Charlie Parker and Diz in the 40s")
    //then
    result := <-releaser
    assert.IsNotNil(result).AreEqual("Louis Armstrong was great!", result["rants-about-fusion"]).
        AreEqual("The best duo was Charlie Parker and Diz in the 40s", result["rants-about-bebop"])
}

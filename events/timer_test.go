package events

import (
    // "testing"
    // "github.com/tholowka/testing/assertions"
    // "time"
)

// func TestThat_WhenTimeout_Works(t *testing.T) {
//     //given
//     assert := assertions.New(t)
//     topic := NewTopic("a-game-to-play")
//     publisher := topic.NewPublisher()
//     waitForAnswer := make(chan bool)
//     //when
//     errorTopic := WhenTimeout(topic, time.Duration(50)*time.Millisecond, "timeouts")
//     errorTopic.NewSubscriber(func(err interface{}) {
//         switch err.(type) {
//         case time.Time:
//             waitForAnswer<-true
//         default:
//             waitForAnswer<-false
//         }
//     })
//     go func() {
//         <-time.After(time.Duration(100)*time.Millisecond)
//         publisher("hello")
//     }()
//     //then
//     assert.IsTrue(<-waitForAnswer)
//     errorTopic.Close()
// }
//
// func TestThat_WhenTimeout_Resets_EachTimeAnEventHappens(t *testing.T) {
//     //given
//     assert := assertions.New(t)
//     <-time.After(time.Duration(2)*time.Second)
//     assert.IsTrue(true)
//     <-time.After(time.Duration(2)*time.Second)
//
//     // topic := NewTopic("a-game-to-play")
//     // publisher := topic.NewPublisher()
//     // waitForAnswer := make(chan time.Time)
//     // //when
//     // errorTopic := WhenTimeout(topic, time.Duration(50)*time.Millisecond, "timeouts")
//     // errorTopic.NewSubscriber(func(err interface{}) {
//     //     switch err.(type) {
//     //     case time.Time:
//     //         waitForAnswer<-err.(time.Time)
//     //     default:
//     //     }
//     // })
//     // startTime := time.Now()
//     // go func() {
//     //     <-time.After(time.Duration(30)*time.Millisecond)
//     //     publisher("hello")
//     // }()
//     // //then
//     // log.Println(startTime)
//     // delay := <-waitForAnswer
//     // log.Println(delay)
//     // assert.IsTrue(false)
//     // errorTopic.Close()
// }

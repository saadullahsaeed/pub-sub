package events

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "time"
    "encoding/json"
    "io/ioutil"
    "errors"
    "fmt"
)

func Test_And_WithMultipleTopics(t *testing.T) {
    //given
    results := runFixtureAndOp("test1.json", And)
    //then
    expectResult(assertions.New(t), <-results, map[string][]interface{} {
        "topic1" : []interface{} { "hello", "how are you?" },
        "topic2" : []interface{} { "hello" },
    })
}

func Test_And_WithMultipleTopics_And_A_MissingPublish(t *testing.T) {
    //given
    results := runFixtureAndOp("test2.json", And)
    //then
    expectError(assertions.New(t), <-results)
}

func Test_And_WithMultipleTopics_And_That_It_DoesntWait_For_Late_Publishes(t *testing.T) {
    //given
    results := runFixtureAndOp("test3.json", And)
    //then
    expectResult(assertions.New(t), <-results, map[string][]interface{} {
        "topic1" : []interface{} { "hello" },
        "topic2" : []interface{} { "don't want to talk to you anymore" },
    })
}

func Test_Close_DoesNot_CrashAnything(t *testing.T) {
    //given
    assert := assertions.New(t)
    rants := NewTopicWithLogging("rants", defaultLogging)
    streams := NewTopicWithLogging("streams", defaultLogging)
    joint := And([]Topic { rants, streams }, "joint")
    //then
    assert.DoesNotThrow(func() {
        joint.Close()
    })
}

type fixture struct {
    Topics []string `json:"topics"`
    Test []map[string]string `json:"test"`
}

func loadFixture(filepath string) *fixture {
    bytes, err := ioutil.ReadFile(filepath)
    if err != nil {
        panic(err.Error())
    }
    testData := &fixture { []string{}, []map[string]string {} }
    if err = json.Unmarshal(bytes, testData); err != nil {
        panic(err.Error())
    }
    return testData
}

func runFixtureAndOp(filepath string, topicOperation func([]Topic, string) Topic) chan interface{} {
    fixture := loadFixture(filepath)
    topics := map[string]Topic {}
    topicsArray := []Topic {}
    results := make(chan interface{})
    for _, name := range fixture.Topics {
        topics[name] = NewTopicWithLogging(name, defaultLogging)
        topicsArray = append(topicsArray, topics[name])
    }
    topic := topicOperation(topicsArray, "results")
    for _, publish := range fixture.Test {
        for name, message := range publish {
            //note: the below technique streamlines execution of Publishers...
            //otherwise a publish message later in the Json might be executed earlier than a different one
            streamLine := make(chan bool)
            topics[name].NewSubscriber(func(_ interface{}) {
                streamLine<-true
            })
            topics[name].NewPublisher()(message)
            <-streamLine
        }
    }
    subscriber := func(topicMessage interface{}) {
        results <- topicMessage
    }
    topic.NewSubscriber(subscriber)
    finalResult := make(chan interface{})
    select {
    case message:=<-results:
        go func() {
            finalResult<-message
        }()
    case <-time.After(time.Duration(1)*time.Second):
        go func() {
            finalResult<-errors.New("No message received")
        }()
    }
    go func() {
        //not very effective, but we dont want to use TemporaryTopic here as not to mix dependencies
        <-time.After(time.Second)
        for _, partialTopic := range topicsArray {
            partialTopic.Close()
        }
        topic.Close()
    }()
    return finalResult
}

func expectResult(assert assertions.Assertions, results interface{}, expected interface{}) {
    switch results.(type) {
    case map[string][]interface{}:
        assert.AreEqual(len(expected.(map[string][]interface{})), len(results.(map[string][]interface{})))
        for key, value := range expected.(map[string][]interface{}) {
            assert.AreEqual(value, results.(map[string][]interface{})[key])
        }
    case error:
        defaultLogging("%v",results.(error))
        assert.IsTrue(false)
    default:
        defaultLogging(fmt.Sprintf("Expecting a different type: %T", results))
        assert.IsTrue(false)
    }
}

func expectError(assert assertions.Assertions, results interface{}) {
    switch results.(type) {
    case error:
        assert.IsTrue(true)
    case map[string][]string:
        assert.IsTrue(false)
    default:
        defaultLogging(fmt.Sprintf("Expecting a different type: %T", results))
        assert.IsTrue(false)
    }
}

package events
import (
)

type provider struct {
}

func (t *provider) NewTopic(topicName string) Topic {
    return nil
}

func (t *provider) NewTopicWithLogging(topicName string, loggingMethod func(string, ...interface{})) Topic {
    return nil
}

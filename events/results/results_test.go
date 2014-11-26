package results

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "log"
)

func TestThat_First_ThrowsErrors(t *testing.T) {
    //given
    returnsAnError(New(aString).First, t)
    doesNotReturnAnError(New(aStringArray).First, t)
}

func returnsAnError(method func() (interface{}, error), t *testing.T) {
    result, err := method()
    log.Println(err)
    assertions.New(t).IsNil(result).IsNotNil(err)
}

func doesNotReturnAnError(method func() (interface{}, error), t *testing.T) {
    result, err := method()
    log.Println(err)
    assertions.New(t).IsNotNil(result).IsNil(err)
}

func aString() interface{} {
    return "mingus"
}

func aStringArray() interface{} {
    return []string { "mingus", "ella" }
}

func anInt() interface{} {
    return 0
}

package results

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    "log"
)

const (
    HELLO = "hello"
)

func Test_ArrayBehaviour(t *testing.T) {
    returnsAnError(New(aString()).First, t)
    returnsAnError(New(aStringArray()).First, t)
    doesNotReturnAnError(New(anArrayOfStrings()).First, t)

    returnsAnError(New(aString()).Last, t)
    returnsAnError(New(aStringArray()).Last, t)
    doesNotReturnAnError(New(anArrayOfStrings()).Last, t)
}

func Test_MapBehaviour(t *testing.T) {
    returnsAnError(New(aString()).For(HELLO), t)
    returnsAnError(New(aStringArray()).For(HELLO), t)
    doesNotReturnAnError(New(aMapOfArrays()).For(HELLO), t)
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

func anArrayOfStrings() interface{} {
    return []interface{} { "mingus", "ella" }
}

func aMapOfArrays() interface{} {
    return map[string][]interface{} { HELLO : anArrayOfStrings() }
}


package results

import (
    "testing"
    "github.com/tholowka/testing/assertions"
    // "log"
)

const (
    HELLO = "hello"
)

func Test_ArrayBehaviour(t *testing.T) {
    assert := assertions.New(t)

    assert.IsNotNil(New(aString()).First().Error()).
        IsNotNil(New(aStringArray()).First().Error()).
        IsNil(New(anArrayOfStrings()).First().Error())

    assert.IsNotNil(New(aString()).Last().Error()).
        IsNotNil(New(aStringArray()).Last().Error()).
        IsNil(New(anArrayOfStrings()).Last().Error())
}

func Test_MapBehaviour(t *testing.T) {
    assert := assertions.New(t)

    assert.IsNotNil(New(aString()).For(HELLO).Error()).
        IsNotNil(New(aStringArray()).For(HELLO).Error()).
        IsNil(New(aMapOfArrays()).For(HELLO).Error())
}

func Test_Stacking(t *testing.T) {
    assert := assertions.New(t)

    assert.IsNil(New(aMapOfArrays()).For(HELLO).First().Error())
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
    return map[string]interface{} { HELLO : anArrayOfStrings() }
}


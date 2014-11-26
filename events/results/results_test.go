package results

import (
    "testing"
    "github.com/tholowka/testing/assertions"
)

func TestThat_First_ThrowsErrors(t *testing.T) {
    //given
    returnsAnError(New(aString).First, t)

}

func returnsAnError(method func() (interface{}, error), t *testing.T) {
    result, err := method()
    assertions.New(t).IsNil(result).IsNotNil(err)
}

func doesNotReturnAnError(method func() (interface{}, error), t *testing.T) {
    result, err := method()
    assertions.New(t).IsNotNil(result).IsNil(err)
}

func aString() interface{} {
    return "mingus"
}

func anInt() interface{} {
    return 0
}

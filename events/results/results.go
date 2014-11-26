/**
This package offers some casting and flattening logic which can be useful when you start dealing with compound Topic results. 

Topic aggregation functions like And and Or provide you with events that have the form map[string][]interface{}. Quite often, 
the interfaces used in the values are themselves arrays or maps as well. This makes transformations tedious and problematic. 

For example, within a call to NewSubscriber(func(event interface{}) you could use this snippet of code:

result, err := New(event).First to obtain the first element in the array, or an error if the event is not []interface{}

Many of the methods here have their XXXAndWrap variant, which does not return the interface{}, but 'wraps' the interface{} with a call 
to New. This allows for a Fluent API approach can be taken.
*/
package results

import (
    "errors"
    "fmt"
)

/* A wrapper around a Topic result which allows for casting and flattening functionalities 
*/
type Results struct {
    result interface{}
    err error
}

func New(result interface{}) *Results {
    return &Results { result, nil }
}

func from(err error) *Results {
    return &Results { nil, err }
}

func (r *Results) First() *Results {
    return r.At(0)
}

func (r *Results) Last() *Results {
    if r.err != nil {
        return from(r.err)
    }
    switch r.result.(type) {
    case []interface{}:
        result := r.result.([]interface{})
        if len(result) > 0 {
            return New(result[len(result)-1])
        } else {
            return nil
        }
    default:
    }
    return from(r.toError("This is not an array of interfaces"))
}

func (r *Results) At(index int) *Results {
    if r.err != nil {
        return from(r.err)
    }
    switch r.result.(type) {
    case []interface{}:
        return New(r.result.([]interface{})[index])
    default:
    }
    return from(r.toError("This is not an array of interfaces"))
}

func (r *Results) For(key string) *Results {
    if r.err != nil {
        return from(r.err)
    }
    switch r.result.(type) {
    case map[string]interface{}:
        return New(r.result.(map[string]interface{})[key])
    default:
    }
    return from(r.toError("This is not a map"))
}

func (r *Results) Invoke(callback func(interface{})) {
    callback(r.result)
}

func (r *Results) Get() interface{} {
    return r.result
}

func (r *Results) Error() error {
    return r.err
}

func (r *Results) toError(message string) error {
    return errors.New(fmt.Sprintf(message+": %v %T", r.result, r.result))
}


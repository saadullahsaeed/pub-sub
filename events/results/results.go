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
}

func New(result interface{}) *Results {
    return &Results { result }
}

func (r *Results) FirstAndWrap() (*Results, error) {
    return r.AtAndWrap(0)
}

func (r *Results) First() (interface{}, error) {
    return r.At(0)
}

func (r *Results) LastAndWrap() (*Results, error) {
    switch r.result.(type) {
    case []interface{}:
        result := r.result.([]interface{})
        if len(result) > 0 {
            return New(result[len(result)-1]), nil
        } else {
            return nil, nil
        }
    default:
    }
    return nil, r.toError("This is not an array of interfaces")
}

func (r *Results) Last() (interface{}, error) {
    result, err := r.LastAndWrap()
    return r.cast(result, err)
}

func (r *Results) AtAndWrap(index int) (*Results, error) {
    switch r.result.(type) {
    case []interface{}:
        return New(r.result.([]interface{})[index]), nil
    default:
    }
    return nil, r.toError("This is not an array of interfaces")
}

func (r *Results) At(index int) (interface{}, error) {
    result, err := r.AtAndWrap(index)
    return r.cast(result, err)
}

func (r *Results) ForAndWrap(key string) (*Results, error) {
    switch r.result.(type) {
    case map[string]interface{}:
        return New(r.result.(map[string]interface{})[key]), nil
    default:
    }
    return nil, r.toError("This is not a map")

}

func (r *Results) For(key string) (interface{}, error) {
    result, err := r.ForAndWrap(key)
    return r.cast(result, err)
}

func (r *Results) Invoke(callback func(interface{})) {
    callback(r.result)
}

func (r *Results) Get() interface{} {
    return r.result
}

func (r *Results) cast(result *Results, err error) (interface{}, error) {
    if result != nil {
        return result.Get, err
    } else {
        return nil, err
    }
}

func (r *Results) toError(message string) error {
    return errors.New(fmt.Sprintf(message+": %v %T", r.result, r.result))
}


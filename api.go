package services

import (
	"fmt"
	"github.com/kiprotect/go-helpers/forms"
	"reflect"
)

type Context interface {
	Params() map[string]interface{}
	Result(data interface{}) Response
	Error(code int, message string, data interface{}) Response
	InternalError() Response
	InvalidParams(err error) Response
	MethodNotFound() Response
	NotFound() Response
	Acknowledge() Response
	Nil() Response
}

type Response interface {
}

type Request interface {
}

// calls the handler with the validated and coerced form parameters
func callHandler(context Context, handler, params interface{}) (Response, error) {
	value := reflect.ValueOf(handler)

	if value.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	paramsValue := reflect.ValueOf(params)
	contextValue := reflect.ValueOf(context)

	responseValue := value.Call([]reflect.Value{contextValue, paramsValue})

	return responseValue[0].Interface().(Response), nil

}

func HandleAPICall(handler interface{}, form *forms.Form, context Context) Response {
	if params, err := form.ValidateWithContext(context.Params(), map[string]interface{}{"context": context}); err != nil {
		return context.InvalidParams(err)
	} else {
		if paramsStruct, err := APIHandlerStruct(handler); err != nil {
			// this should never happen...
			Log.Error(err)
			return context.InternalError()
		} else if err := form.Coerce(paramsStruct, params); err != nil {
			// this shouldn't happen either...
			Log.Error(err)
			return context.InternalError()
		} else {
			if response, err := callHandler(context, handler, paramsStruct); err != nil {
				// and neither should this...
				Log.Error(err)
				return context.InternalError()
			} else {
				if response == nil {
					return context.Nil()
				}
				return response
			}
		}
	}

}

// returns a new struct that we can coerce the valid form parameters into
func APIHandlerStruct(handler interface{}) (interface{}, error) {
	value := reflect.ValueOf(handler)
	if value.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	funcType := value.Type()

	if funcType.NumIn() != 2 {
		return nil, fmt.Errorf("expected a function with 2 arguments")
	}

	if funcType.NumOut() != 1 {
		return nil, fmt.Errorf("expected a function with 1 return value")
	}

	returnType := funcType.Out(0)

	if !returnType.Implements(reflect.TypeOf((*Response)(nil)).Elem()) {
		return nil, fmt.Errorf("return value should be a response")
	}

	contextType := funcType.In(0)

	if !contextType.Implements(reflect.TypeOf((*Context)(nil)).Elem()) {
		return nil, fmt.Errorf("first argument should accept a context")
	}

	structType := funcType.In(1)

	if structType.Kind() != reflect.Ptr || structType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("second argument should be a struct pointer")
	}

	// we create a new struct and return it
	return reflect.New(structType.Elem()).Interface(), nil
}

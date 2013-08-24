package filter

import (
	"code.google.com/p/go.net/websocket"
	"github.com/robfig/revel"
	"reflect"
	"runtime"
	"strings"
)

var (
	controllerFilters map[reflect.Type][]*RegisteredMethod = make(map[reflect.Type][]*RegisteredMethod)
)

type RegisteredMethod struct {
	When             revel.When
	Methods          []string //registered methods
	TargetMethod     interface{}
	TargetMethodName string
}

func AddControllerFilter(target interface{}, when revel.When, methods ...string) {
	fullName := runtime.FuncForPC(reflect.ValueOf(target).Pointer()).Name()
	tokens := strings.Split(fullName, ".")
	methodName := tokens[len(tokens)-1] // filter method name in string dt

	receiverType := reflect.TypeOf(target).In(0) // the receiver type is actually the controller static type

	controllerFilters[receiverType] = append(controllerFilters[receiverType],
		&RegisteredMethod{When: when, TargetMethod: target, TargetMethodName: methodName, Methods: methods})
}

func ControllerFilter(c *revel.Controller, fc []revel.Filter) {

	// The receiver of the filter method, should be the controller instance
	receiver := reflect.ValueOf(c.AppController).Elem()

	var resultValuesBefore []reflect.Value
	// Call before
	for _, registeredMethod := range controllerFilters[c.Type.Type] {
		if registeredMethod.When == revel.BEFORE {
			resultValuesBefore = append(resultValuesBefore, invokeMethod(receiver, registeredMethod, c))
		}
	}
	resultValue := getResultValue(resultValuesBefore)

	// The filter chain only continue when the result Value is nil
	if !resultValue.IsValid() || resultValue.IsNil() {
		fc[0](c, fc[1:])

		var resultValuesAfter []reflect.Value
		// Call after
		for _, registeredMethod := range controllerFilters[c.Type.Type] {
			if registeredMethod.When == revel.AFTER {
				resultValuesAfter = append(resultValuesAfter, invokeMethod(receiver, registeredMethod, c))
			}
		}
		resultValue = getResultValue(resultValuesAfter)
	}

	// only set the c.Result if the Action should not be called or it returns nil
	if c.Result == nil && resultValue.Kind() == reflect.Interface && !resultValue.IsNil() {
		c.Result = resultValue.Interface().(revel.Result)
	}
}

// bind the parameter to the filter methods based on the method definition
func bindParameter(receiver reflect.Value, methodType *revel.MethodType, params *revel.Params) []reflect.Value {
	var methodArgs []reflect.Value
	methodArgs = append(methodArgs, receiver)

	// Bind the funciton signature
	for _, arg := range methodType.Args {
		var boundArg reflect.Value
		// Ignore websocket for now
		if arg.Type != reflect.TypeOf((*websocket.Conn)(nil)) {
			boundArg = revel.Bind(params, arg.Name, arg.Type)
		}

		methodArgs = append(methodArgs, boundArg)
	}

	return methodArgs
}

// invoke the filter method and get the result
func invokeMethod(receiver reflect.Value, registeredMethod *RegisteredMethod, c *revel.Controller) (resultValue reflect.Value) {
	for _, method := range registeredMethod.Methods {
		if strings.EqualFold(method, c.MethodName) {
			methodType := c.Type.Method(registeredMethod.TargetMethodName)
			targetMethod := reflect.ValueOf(registeredMethod.TargetMethod)

			methodArgs := bindParameter(receiver, methodType, c.Params)
			resultValue = targetMethod.Call(methodArgs)[0]
			return
		}
	}
	return
}

// get the first valid result value as the final value for each When, if there are multiples
func getResultValue(resultValues []reflect.Value) (resultValue reflect.Value) {
	for _, resultValue = range resultValues {
		if resultValue.IsValid() && !resultValue.IsNil() {
			return
		}
	}
	return
}

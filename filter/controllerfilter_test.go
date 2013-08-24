package filter

import (
	"fmt"
	"github.com/robfig/revel"
	"net/url"
	"reflect"
	"testing"
)

type TestController struct {
	*revel.Controller
}

func (t TestController) Show(id string, x int) revel.Result {
	fmt.Println("show sth")
	return t.Redirect("/")
}

func (t TestController) Edit(id string, x int) revel.Result {
	fmt.Println("edit sth")
	return t.Redirect("/")
}

func (t TestController) isOwner(id string) revel.Result {
	fmt.Println("before, really works:" + id)
	return nil // The filter chain only continue if the return value is nil
}

func (t TestController) callAfter(id string, x int) revel.Result {
	fmt.Println("after, really works:" + id)
	fmt.Println(x)
	return nil // should be always nil, otherwise it will override the value set in Action
}

func TestAddControllerFilter(t *testing.T) {
	AddControllerFilter(TestController.isOwner, revel.BEFORE, "Show", "Edit")
	AddControllerFilter(TestController.callAfter, revel.AFTER, "Show")

	c := revel.NewController(nil, nil)
	controller := &TestController{}
	c.AppController = controller
	c.MethodName = "Show"
	c.Params = &revel.Params{Values: make(url.Values)}
	c.Params.Set("id", "cool")
	c.Params.Set("x", "5")

	var methods []*revel.MethodType
	methods = append(methods,
		&revel.MethodType{
			Name: "isOwner",
			Args: []*revel.MethodArg{&revel.MethodArg{Name: "id", Type: reflect.TypeOf((*string)(nil))}}})
	methods = append(methods,
		&revel.MethodType{
			Name: "callAfter",
			Args: []*revel.MethodArg{
				&revel.MethodArg{Name: "id", Type: reflect.TypeOf((*string)(nil))},
				&revel.MethodArg{Name: "x", Type: reflect.TypeOf((*int)(nil))}}})
	revel.RegisterController(c, methods)

	c.Type = &revel.ControllerType{Type: reflect.TypeOf(controller).Elem(), Methods: methods}

	// fmt.Println(c.Type.Method("isOwner"))

	methodArg := []*revel.MethodArg{
		&revel.MethodArg{Name: "id", Type: reflect.TypeOf("")},
		&revel.MethodArg{Name: "x", Type: reflect.TypeOf(1)}}
	c.MethodType = &revel.MethodType{Name: "Show", Args: methodArg}

	flt := func(c *revel.Controller, fc []revel.Filter) {}

	fc := []revel.Filter{flt}
	fmt.Println("Going to call Show:")
	ControllerFilter(c, fc)

	//test Edit method
	c.MethodName = "Edit"
	c.MethodType = &revel.MethodType{Name: "Edit", Args: methodArg}
	fmt.Println("Going to call Edit:")
	ControllerFilter(c, fc)
}

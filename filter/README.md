Parametered Controller Filter for Revel webframework
----------------

With this Revel fitler, make it possible to define controller method that can be triggerred BEFORE
or AFTER the Action be called, with all the parameters ready for use.

The method will be binding with the parameters according to its signature. The signature of the filter method could 
be different from defined on the Action.

If you have a 'generic' method want to register for different controller & specific methods, it is achievable
via define this method on a 'parent' controller and define methods on each 'child' controller to call it, then 
register the methods on 'child' for each controller.

### How to install 
`go get github.com/arkxu/revel.filter`

### How to use

First of all, it is required to remove the following lines in reflect.go in revel project
(https://github.com/robfig/revel/blob/master/harness/reflect.go#L400).

```Go
  // Is it public?
  // if !funcDecl.Name.IsExported() {
  //  return
  // }
```

This is to allow revel to register both public and private Actions (all of them returns revel.Result which is a balanced restriction).
Those are actually the filter methods.

Then it is required to run

```Bash
go get github.com/robfig/revel/revel
```

So that it recompiles the revel.


### Sample

In the sample, the method `isOwner` will be called before `Edit`, `Delete` or `Update` a question, passing the question id
or the Question instance. 
The method `callAfter` will be called after `Show` a question.

Please check the test code for more detailed info.

Register the filter:

```Go
  package app

  import (
    "github.com/arkxu/revel.filter/filter"
    "github.com/robfig/revel"
  )

  func init() {
    // Filters is the default set of global filters.
    revel.Filters = []revel.Filter{
      revel.PanicFilter,             // Recover from panics and display an error page instead.
      revel.RouterFilter,            // Use the routing table to select the right Action
      revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
      revel.ParamsFilter,            // Parse parameters into Controller.Params.
      revel.SessionFilter,           // Restore and write the session cookie.
      revel.FlashFilter,             // Restore and write the flash cookie.
      revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
      revel.I18nFilter,              // Resolve the requested language
      revel.InterceptorFilter,       // Run interceptors around the action.
      filter.ControllerFilter, // Parametered controller action, should be put just before ActionInvoker
      revel.ActionInvoker, // Invoke the action.
    }
  }
```

Add then register the method need to be called BEFORE or AFTER:

```Go
  package controllers

  import (
    "github.com/arkxu/revel.filter/filter"
    "github.com/robfig/revel"
  )

  type Questions struct {
    *revel.Controller
  }


  func (c Questions) Edit(id string) revel.Result {
    // created a question here
    return c.Redirect("/questions")
  }

  func (c Questions) Delete(id string) revel.Result {
    // deleted a question here
    return c.Redirect("/questions")
  }

  func (c Questions) Update(question *model.Question) revel.Result {
    // deleted a question here
    return c.Redirect("/questions")
  }

  // GET /questions/1
  func (c Questions) Show(id string) revel.Result {
    question := "fetched a question"
    return c.Render(question)
  }

  func (c Questions) isOwner(id string) revel.Result {
    if id != "123"{
      return c.Redirect("/login") // will not invoke the Action, instead it will redirect to login
    }
    return nil //return nil will continue invoke the Action
  }

  func (c Questions) callAfter(id string) revel.Result {
    // do something
    return nil //should always be nil for AFTER methods, unless you really know what you are doing here
  }

  func init() {
    filter.AddControllerFilter(Questions.isOwner, revel.BEFORE, "Edit", "Delete", "Update")
    filter.AddControllerFilter(Questions.callAfter, revel.AFTER, "Show")
  }

```
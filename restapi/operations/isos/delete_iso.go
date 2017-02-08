package isos

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// DeleteIsoHandlerFunc turns a function with the right signature into a delete iso handler
type DeleteIsoHandlerFunc func(DeleteIsoParams) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteIsoHandlerFunc) Handle(params DeleteIsoParams) middleware.Responder {
	return fn(params)
}

// DeleteIsoHandler interface for that can handle valid delete iso params
type DeleteIsoHandler interface {
	Handle(DeleteIsoParams) middleware.Responder
}

// NewDeleteIso creates a new http.Handler for the delete iso operation
func NewDeleteIso(ctx *middleware.Context, handler DeleteIsoHandler) *DeleteIso {
	return &DeleteIso{Context: ctx, Handler: handler}
}

/*DeleteIso swagger:route DELETE /isos/{name} Isos deleteIso

Delete Iso

*/
type DeleteIso struct {
	Context *middleware.Context
	Handler DeleteIsoHandler
}

func (o *DeleteIso) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, _ := o.Context.RouteInfo(r)
	var Params = NewDeleteIsoParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}

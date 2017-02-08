package isos

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewGetIsoParams creates a new GetIsoParams object
// with the default values initialized.
func NewGetIsoParams() *GetIsoParams {
	var ()
	return &GetIsoParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetIsoParamsWithTimeout creates a new GetIsoParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetIsoParamsWithTimeout(timeout time.Duration) *GetIsoParams {
	var ()
	return &GetIsoParams{

		timeout: timeout,
	}
}

// NewGetIsoParamsWithContext creates a new GetIsoParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetIsoParamsWithContext(ctx context.Context) *GetIsoParams {
	var ()
	return &GetIsoParams{

		Context: ctx,
	}
}

/*GetIsoParams contains all the parameters to send to the API endpoint
for the get iso operation typically these are written to a http.Request
*/
type GetIsoParams struct {

	/*Name*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get iso params
func (o *GetIsoParams) WithTimeout(timeout time.Duration) *GetIsoParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get iso params
func (o *GetIsoParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get iso params
func (o *GetIsoParams) WithContext(ctx context.Context) *GetIsoParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get iso params
func (o *GetIsoParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithName adds the name to the get iso params
func (o *GetIsoParams) WithName(name string) *GetIsoParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the get iso params
func (o *GetIsoParams) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *GetIsoParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	r.SetTimeout(o.timeout)
	var res []error

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

package isos

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewPostIsoParams creates a new PostIsoParams object
// with the default values initialized.
func NewPostIsoParams() *PostIsoParams {
	var ()
	return &PostIsoParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewPostIsoParamsWithTimeout creates a new PostIsoParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewPostIsoParamsWithTimeout(timeout time.Duration) *PostIsoParams {
	var ()
	return &PostIsoParams{

		timeout: timeout,
	}
}

// NewPostIsoParamsWithContext creates a new PostIsoParams object
// with the default values initialized, and the ability to set a context for a request
func NewPostIsoParamsWithContext(ctx context.Context) *PostIsoParams {
	var ()
	return &PostIsoParams{

		Context: ctx,
	}
}

/*PostIsoParams contains all the parameters to send to the API endpoint
for the post iso operation typically these are written to a http.Request
*/
type PostIsoParams struct {

	/*Body*/
	Body io.ReadCloser
	/*Name*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the post iso params
func (o *PostIsoParams) WithTimeout(timeout time.Duration) *PostIsoParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the post iso params
func (o *PostIsoParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the post iso params
func (o *PostIsoParams) WithContext(ctx context.Context) *PostIsoParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the post iso params
func (o *PostIsoParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithBody adds the body to the post iso params
func (o *PostIsoParams) WithBody(body io.ReadCloser) *PostIsoParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the post iso params
func (o *PostIsoParams) SetBody(body io.ReadCloser) {
	o.Body = body
}

// WithName adds the name to the post iso params
func (o *PostIsoParams) WithName(name string) *PostIsoParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the post iso params
func (o *PostIsoParams) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *PostIsoParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	r.SetTimeout(o.timeout)
	var res []error

	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

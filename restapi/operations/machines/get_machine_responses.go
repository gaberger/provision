package machines

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/galthaus/swagger-test/models"
)

/*GetMachineOK get machine o k

swagger:response getMachineOK
*/
type GetMachineOK struct {

	/*
	  In: Body
	*/
	Payload GetMachineOKBody `json:"body,omitempty"`
}

// NewGetMachineOK creates GetMachineOK with default headers values
func NewGetMachineOK() *GetMachineOK {
	return &GetMachineOK{}
}

// WithPayload adds the payload to the get machine o k response
func (o *GetMachineOK) WithPayload(payload GetMachineOKBody) *GetMachineOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get machine o k response
func (o *GetMachineOK) SetPayload(payload GetMachineOKBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMachineOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}

}

/*GetMachineUnauthorized get machine unauthorized

swagger:response getMachineUnauthorized
*/
type GetMachineUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Result `json:"body,omitempty"`
}

// NewGetMachineUnauthorized creates GetMachineUnauthorized with default headers values
func NewGetMachineUnauthorized() *GetMachineUnauthorized {
	return &GetMachineUnauthorized{}
}

// WithPayload adds the payload to the get machine unauthorized response
func (o *GetMachineUnauthorized) WithPayload(payload *models.Result) *GetMachineUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get machine unauthorized response
func (o *GetMachineUnauthorized) SetPayload(payload *models.Result) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMachineUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetMachineNotFound get machine not found

swagger:response getMachineNotFound
*/
type GetMachineNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Result `json:"body,omitempty"`
}

// NewGetMachineNotFound creates GetMachineNotFound with default headers values
func NewGetMachineNotFound() *GetMachineNotFound {
	return &GetMachineNotFound{}
}

// WithPayload adds the payload to the get machine not found response
func (o *GetMachineNotFound) WithPayload(payload *models.Result) *GetMachineNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get machine not found response
func (o *GetMachineNotFound) SetPayload(payload *models.Result) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMachineNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetMachineInternalServerError get machine internal server error

swagger:response getMachineInternalServerError
*/
type GetMachineInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Result `json:"body,omitempty"`
}

// NewGetMachineInternalServerError creates GetMachineInternalServerError with default headers values
func NewGetMachineInternalServerError() *GetMachineInternalServerError {
	return &GetMachineInternalServerError{}
}

// WithPayload adds the payload to the get machine internal server error response
func (o *GetMachineInternalServerError) WithPayload(payload *models.Result) *GetMachineInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get machine internal server error response
func (o *GetMachineInternalServerError) SetPayload(payload *models.Result) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMachineInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

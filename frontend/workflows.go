package frontend

import (
	"github.com/VictorLowther/jsonpatch2"
	"github.com/digitalrebar/provision/backend"
	"github.com/digitalrebar/provision/models"
	"github.com/gin-gonic/gin"
)

// WorkflowResponse returned on a successful GET, PUT, PATCH, or POST of a single workflow
// swagger:response
type WorkflowResponse struct {
	// in: body
	Body *models.Workflow
}

// WorkflowsResponse returned on a successful GET of all the workflows
// swagger:response
type WorkflowsResponse struct {
	//in: body
	Body []*models.Workflow
}

// WorkflowBodyParameter used to inject a Workflow
// swagger:parameters createWorkflow putWorkflow
type WorkflowBodyParameter struct {
	// in: body
	// required: true
	Body *models.Workflow
}

// WorkflowPatchBodyParameter used to patch a Workflow
// swagger:parameters patchWorkflow
type WorkflowPatchBodyParameter struct {
	// in: body
	// required: true
	Body jsonpatch2.Patch
}

// WorkflowPathParameter used to name a Workflow in the path
// swagger:parameters putWorkflows getWorkflow putWorkflow patchWorkflow deleteWorkflow headWorkflow
type WorkflowPathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
}

// WorkflowListPathParameter used to limit lists of Workflow by path options
// swagger:parameters listWorkflows listStatsWorkflows
type WorkflowListPathParameter struct {
	// in: query
	Offest int `json:"offset"`
	// in: query
	Limit int `json:"limit"`
	// in: query
	Available string
	// in: query
	Valid string
	// in: query
	ReadOnly string
	// in: query
	Name string
	// in: query
	Reboot string
	// in: query
	BootEnv string
}

// WorkflowActionsPathParameter used to find a Workflow / Actions in the path
// swagger:parameters getWorkflowActions
type WorkflowActionsPathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
	// in: query
	Plugin string `json:"plugin"`
}

// WorkflowActionPathParameter used to find a Workflow / Action in the path
// swagger:parameters getWorkflowAction
type WorkflowActionPathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
	// in: path
	// required: true
	Cmd string `json:"cmd"`
	// in: query
	Plugin string `json:"plugin"`
}

// WorkflowActionBodyParameter used to post a Workflow / Action in the path
// swagger:parameters postWorkflowAction
type WorkflowActionBodyParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
	// in: path
	// required: true
	Cmd string `json:"cmd"`
	// in: query
	Plugin string `json:"plugin"`
	// in: body
	// required: true
	Body map[string]interface{}
}

func (f *Frontend) InitWorkflowApi() {
	// swagger:route GET /workflows Workflows listWorkflows
	//
	// Lists Workflows filtered by some parameters.
	//
	// This will show all Workflows by default.
	//
	// You may specify:
	//    Offset = integer, 0-based inclusive starting point in filter data.
	//    Limit = integer, number of items to return
	//
	// Functional Indexs:
	//    Name = string
	//    Reboot = boolean
	//    BootEnv = string
	//    Available = boolean
	//
	// Functions:
	//    Eq(value) = Return items that are equal to value
	//    Lt(value) = Return items that are less than value
	//    Lte(value) = Return items that less than or equal to value
	//    Gt(value) = Return items that are greater than value
	//    Gte(value) = Return items that greater than or equal to value
	//    Between(lower,upper) = Return items that are inclusively between lower and upper
	//    Except(lower,upper) = Return items that are not inclusively between lower and upper
	//
	// Example:
	//    Name=fred - returns items named fred
	//    Name=Lt(fred) - returns items that alphabetically less than fred.
	//    Name=Lt(fred)&Available=true - returns items with Name less than fred and Available is true
	//
	// Responses:
	//    200: WorkflowsResponse
	//    401: NoContentResponse
	//    403: NoContentResponse
	//    406: ErrorResponse
	f.ApiGroup.GET("/workflows",
		func(c *gin.Context) {
			f.List(c, &backend.Workflow{})
		})

	// swagger:route HEAD /workflows Workflows listStatsWorkflows
	//
	// Stats of the List Workflows filtered by some parameters.
	//
	// This will return headers with the stats of the list.
	//
	// You may specify:
	//    Offset = integer, 0-based inclusive starting point in filter data.
	//    Limit = integer, number of items to return
	//
	// Functional Indexs:
	//    Name = string
	//    Reboot = boolean
	//    BootEnv = string
	//    Available = boolean
	//
	// Functions:
	//    Eq(value) = Return items that are equal to value
	//    Lt(value) = Return items that are less than value
	//    Lte(value) = Return items that less than or equal to value
	//    Gt(value) = Return items that are greater than value
	//    Gte(value) = Return items that greater than or equal to value
	//    Between(lower,upper) = Return items that are inclusively between lower and upper
	//    Except(lower,upper) = Return items that are not inclusively between lower and upper
	//
	// Example:
	//    Name=fred - returns items named fred
	//    Name=Lt(fred) - returns items that alphabetically less than fred.
	//    Name=Lt(fred)&Available=true - returns items with Name less than fred and Available is true
	//
	// Responses:
	//    200: NoContentResponse
	//    401: NoContentResponse
	//    403: NoContentResponse
	//    406: ErrorResponse
	f.ApiGroup.HEAD("/workflows",
		func(c *gin.Context) {
			f.ListStats(c, &backend.Workflow{})
		})

	// swagger:route POST /workflows Workflows createWorkflow
	//
	// Create a Workflow
	//
	// Create a Workflow from the provided object
	//
	//     Responses:
	//       201: WorkflowResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.POST("/workflows",
		func(c *gin.Context) {
			b := &backend.Workflow{}
			f.Create(c, b)
		})
	// swagger:route GET /workflows/{name} Workflows getWorkflow
	//
	// Get a Workflow
	//
	// Get the Workflow specified by {name} or return NotFound.
	//
	//     Responses:
	//       200: WorkflowResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/workflows/:name",
		func(c *gin.Context) {
			f.Fetch(c, &backend.Workflow{}, c.Param(`name`))
		})

	// swagger:route HEAD /workflows/{name} Workflows headWorkflow
	//
	// See if a Workflow exists
	//
	// Return 200 if the Workflow specifiec by {name} exists, or return NotFound.
	//
	//     Responses:
	//       200: NoContentResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: NoContentResponse
	f.ApiGroup.HEAD("/workflows/:name",
		func(c *gin.Context) {
			f.Exists(c, &backend.Workflow{}, c.Param(`name`))
		})

	// swagger:route PATCH /workflows/{name} Workflows patchWorkflow
	//
	// Patch a Workflow
	//
	// Update a Workflow specified by {name} using a RFC6902 Patch structure
	//
	//     Responses:
	//       200: WorkflowResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       406: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.PATCH("/workflows/:name",
		func(c *gin.Context) {
			f.Patch(c, &backend.Workflow{}, c.Param(`name`))
		})

	// swagger:route PUT /workflows/{name} Workflows putWorkflow
	//
	// Put a Workflow
	//
	// Update a Workflow specified by {name} using a JSON Workflow
	//
	//     Responses:
	//       200: WorkflowResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.PUT("/workflows/:name",
		func(c *gin.Context) {
			f.Update(c, &backend.Workflow{}, c.Param(`name`))
		})

	// swagger:route DELETE /workflows/{name} Workflows deleteWorkflow
	//
	// Delete a Workflow
	//
	// Delete a Workflow specified by {name}
	//
	//     Responses:
	//       200: WorkflowResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.DELETE("/workflows/:name",
		func(c *gin.Context) {
			f.Remove(c, &backend.Workflow{}, c.Param(`name`))
		})

	workflow := &backend.Workflow{}
	pActions, pAction, pRun := f.makeActionEndpoints(workflow.Prefix(), workflow, "name")

	// swagger:route GET /workflows/{name}/actions Workflows getWorkflowActions
	//
	// List workflow actions Workflow
	//
	// List Workflow actions for a Workflow specified by {name}
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       200: ActionsResponse
	//       401: NoWorkflowResponse
	//       403: NoWorkflowResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/workflows/:name/actions", pActions)

	// swagger:route GET /workflows/{name}/actions/{cmd} Workflows getWorkflowAction
	//
	// List specific action for a workflow Workflow
	//
	// List specific {cmd} action for a Workflow specified by {name}
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       200: ActionResponse
	//       400: ErrorResponse
	//       401: NoWorkflowResponse
	//       403: NoWorkflowResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/workflows/:name/actions/:cmd", pAction)

	// swagger:route POST /workflows/{name}/actions/{cmd} Workflows postWorkflowAction
	//
	// Call an action on the node.
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//
	//     Responses:
	//       400: ErrorResponse
	//       200: ActionPostResponse
	//       401: NoWorkflowResponse
	//       403: NoWorkflowResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	f.ApiGroup.POST("/workflows/:name/actions/:cmd", pRun)
}

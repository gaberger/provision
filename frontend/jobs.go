package frontend

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/VictorLowther/jsonpatch2"
	"github.com/digitalrebar/provision/backend"
	"github.com/digitalrebar/provision/models"
	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
)

// NOTE: Jobs are restricted by Machine UUID

// JobResponse return on a successful GET, PUT, PATCH or POST of a single Job
// swagger:response
type JobResponse struct {
	// in: body
	Body *models.Job
}

// JobsResponse return on a successful GET of all Jobs
// swagger:response
type JobsResponse struct {
	// in: body
	Body []*models.Job
}

// JobActionsResponse return on a successful GET of a Job's actions
// swagger:response
type JobActionsResponse struct {
	// in: body
	Body []*models.JobAction
}

// JobParamsResponse return on a successful GET of all Job's Params
// swagger:response
type JobParamsResponse struct {
	// in: body
	Body map[string]interface{}
}

// This is a HACK - I can't figure out how to get
// swagger to render this a binary.  So we lie.
// We also override this object from the server
// directory to have a binary format which
// turns it into a stream.
//
// JobLogResponse returned on a successful GET of a log
// swagger:response
type JobLogResponse struct {
	// in: body
	// format: binary
	Body string
}

// JobBodyParameter used to inject a Job
// swagger:parameters createJob putJob
type JobBodyParameter struct {
	// in: body
	// required: true
	Body *models.Job
}

// JobPatchBodyParameter used to patch a Job
// swagger:parameters patchJob
type JobPatchBodyParameter struct {
	// in: body
	// required: true
	Body jsonpatch2.Patch
}

// JobLogBodyParameter used to append to a Job log
// swagger:parameters putJobLog
type JobLogPutBodyParameter struct {
	// in: body
	// required: true
	Body interface{}
}

// JobPathParameter used to find a Job in the path
// swagger:parameters putJobs getJob putJob patchJob deleteJob getJobParams postJobParams getJobActions getJobLog putJobLog headJob
type JobPathParameter struct {
	// in: path
	// required: true
	// swagger:strfmt uuid
	Uuid uuid.UUID `json:"uuid"`
}

// JobParamsBodyParameter used to set Job Params
// swagger:parameters postJobParams
type JobParamsBodyParameter struct {
	// in: body
	// required: true
	Body map[string]interface{}
}

// JobListPathParameter used to limit lists of Job by path options
// swagger:parameters listJobs listStatsJobs
type JobListPathParameter struct {
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
	Uuid string
	// in: query
	Stage string
	// in: query
	Task string
	// in: query
	State string
	// in: query
	Machine string
	// in: query
	Archived string
	// in: query
	StartTime string
	// in: query
	EndTime string
}

// JobActionsPathParameter used to find a Job / Actions in the path
// swagger:parameters getJobActions
type JobActionsPathParameter struct {
	// in: path
	// required: true
	// swagger:strfmt uuid
	Uuid uuid.UUID `json:"uuid"`
	// in: query
	Plugin string `json:"plugin"`
}

// JobActionPathParameter used to find a Job / Action in the path
// swagger:parameters getJobAction
type JobActionPathParameter struct {
	// in: path
	// required: true
	// swagger:strfmt uuid
	Uuid uuid.UUID `json:"uuid"`
	// in: path
	// required: true
	Cmd string `json:"cmd"`
	// in: query
	Plugin string `json:"plugin"`
}

// JobActionBodyParameter used to post a Job / Action in the path
// swagger:parameters postJobAction
type JobActionBodyParameter struct {
	// in: path
	// required: true
	// swagger:strfmt uuid
	Uuid uuid.UUID `json:"uuid"`
	// in: path
	// required: true
	Cmd string `json:"cmd"`
	// in: query
	Plugin string `json:"plugin"`
	// in: body
	// required: true
	Body map[string]interface{}
}

func (f *Frontend) InitJobApi() {
	// swagger:route GET /jobs Jobs listJobs
	//
	// Lists Jobs filtered by some parameters.
	//
	// This will show all Jobs by default.
	//
	// You may specify:
	//    Offset = integer, 0-based inclusive starting point in filter data.
	//    Limit = integer, number of items to return
	//
	// Functional Indexs:
	//    Uuid = string
	//    Stage = string
	//    Task = string
	//    State = string
	//    Machine = string
	//    Archived = boolean
	//    StartTime = datetime
	//    EndTime = datetime
	//    Available = boolean
	//    Valid = boolean
	//    ReadOnly = boolean
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
	//    Uuid=fred - returns items named fred
	//    Uuid=Lt(fred) - returns items that alphabetically less than fred.
	//    Uuid=Lt(fred)&Archived=true - returns items with Uuid less than fred and Archived is true
	//
	// Responses:
	//    200: JobsResponse
	//    401: NoContentResponse
	//    403: NoContentResponse
	//    406: ErrorResponse
	f.ApiGroup.GET("/jobs",
		func(c *gin.Context) {
			f.List(c, &backend.Job{})
		})

	// swagger:route HEAD /jobs Jobs listStatsJobs
	//
	// Stats of the List Jobs filtered by some parameters.
	//
	// This will return headers with the stats of the list.
	//
	// You may specify:
	//    Offset = integer, 0-based inclusive starting point in filter data.
	//    Limit = integer, number of items to return
	//
	// Functional Indexs:
	//    Uuid = string
	//    Stage = string
	//    Task = string
	//    State = string
	//    Machine = string
	//    Archived = boolean
	//    StartTime = datetime
	//    EndTime = datetime
	//    Available = boolean
	//    Valid = boolean
	//    ReadOnly = boolean
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
	//    Uuid=fred - returns items named fred
	//    Uuid=Lt(fred) - returns items that alphabetically less than fred.
	//    Uuid=Lt(fred)&Archived=true - returns items with Uuid less than fred and Archived is true
	//
	// Responses:
	//    200: NoContentResponse
	//    401: NoContentResponse
	//    403: NoContentResponse
	//    406: ErrorResponse
	f.ApiGroup.HEAD("/jobs",
		func(c *gin.Context) {
			f.ListStats(c, &backend.Job{})
		})

	// swagger:route POST /jobs Jobs createJob
	//
	// Create a Job
	//
	// Create a Job from the provided object, Only Machine and UUID are used.
	//
	//     Responses:
	//       201: JobResponse
	//       202: JobResponse
	//       204: NoContentResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	//       500: ErrorResponse
	f.ApiGroup.POST("/jobs",
		func(c *gin.Context) {
			// We don't use f.Create() because we need to be able to assign random
			// UUIDs to new Jobs without forcing the client to do so, yet allow them
			// for testing purposes amd if they alrady have a UUID scheme for jobs.
			b := &backend.Job{}
			if !assureDecode(c, b) {
				return
			}
			if b.Machine == nil {
				c.JSON(http.StatusBadRequest, models.NewError(c.Request.Method, http.StatusBadRequest, "Create request must have Machine field"))
				return
			}
			if !f.assureAuth(c, "jobs", "create", b.AuthKey()) {
				return
			}
			if b.Uuid == nil || len(b.Uuid) == 0 {
				b.Uuid = uuid.NewRandom()
			}
			var res models.Model
			var err error
			rt := f.rt(c, b.Locks("create")...)
			var code int
			rt.Do(func(d backend.Stores) {
				cj := backend.ModelToBackend(&models.Job{}).(*backend.Job)
				cj.Uuid = uuid.Parse("00000000-0000-0000-0000-000000000000")
				cj.State = "finished"
				mo := rt.Find("machines", b.Machine.String())
				if mo == nil {
					err = &models.Error{Code: http.StatusUnprocessableEntity, Type: backend.ValidationError,
						Messages: []string{fmt.Sprintf("Machine %s does not exist", b.Machine.String())}}
					code = http.StatusUnprocessableEntity
					return
				}
				oldM := backend.AsMachine(mo)

				// Machine isn't runnable return conflict
				if !(oldM.Runnable && oldM.Available) {
					err = &models.Error{Code: http.StatusConflict, Type: "Conflict",
						Messages: []string{fmt.Sprintf("Machine %s is not runnable", b.Machine.String())}}
					code = http.StatusConflict
					return
				}
				m := backend.ModelToBackend(models.Clone(oldM)).(*backend.Machine)
				m.InRunner()
				defer func() {
					switch code {
					case http.StatusNoContent:
						if m.CurrentTask > len(m.Tasks) {
							m.CurrentTask = len(m.Tasks)
						}
						if oldM.CurrentTask != m.CurrentTask {
							if _, err = rt.Update(m); err != nil {
								code = http.StatusInternalServerError
							}
						}
					}
				}()
				// Are we running a job or not on list yet, do some checking.
				if jo := rt.Find("jobs", m.CurrentJob.String()); jo != nil {
					cj = jo.(*backend.Job)
				}
				code = http.StatusNoContent
				if m.CurrentTask >= len(m.Tasks) {
					return
				}
				advance := true
				if m.CurrentTask == -1 {
					// At this point, we are starting over on the task list
					// We could have been forced and need to close out a job.
					// if it is running, created, or incomplete, we need
					// to mark it failed, but leave us runnable.
					switch cj.State {
					case "running", "created", "incomplete":
						cj.State = "failed"
						if _, err = rt.Update(cj); err != nil {
							code = http.StatusBadRequest
							return
						}
						m.Runnable = true
					}
					m.CurrentTask = 0
					advance = false
				}
				// If we return from inside this for loop, it will be with NoContent
				code = http.StatusNoContent
				for st := strings.SplitN(m.Tasks[m.CurrentTask], ":", 2); len(st) == 2 && m.CurrentTask < len(m.Tasks); m.CurrentTask++ {
					needUpdate := false
					// Handle bootenv and stage changes if needed
					switch st[0] {
					case "stage":
						if m.Stage != st[1] {
							needUpdate = true
							m.Stage = st[1]
						}
					case "bootenv":
						if m.BootEnv != st[1] {
							needUpdate = true
							m.BootEnv = st[1]
						}
					default:
						code = http.StatusInternalServerError
						err = &models.Error{
							Code:  code,
							Type:  "InvalidTaskList",
							Key:   m.Key(),
							Model: m.Prefix(),
						}
						err.(*models.Error).Errorf("Invalid task list entry[%d]: '%s'", m.CurrentTask, m.Tasks[m.CurrentTask])
						return
					}
					if !needUpdate {
						// Already at the right whatever it is, so continue on to try and find a task
						continue
					}
					// We need to update the machine.  Do so, then create a fake job that is already
					// finished to commemorate the occasion, save it, and return
					_, err = rt.Update(m)
					if err != nil {
						code = http.StatusInternalServerError
						return
					}
					b.State = "finished"
					b.StartTime = time.Now()
					b.Previous = cj.Uuid
					b.Machine = m.Uuid
					b.Stage = m.Stage
					b.Task = m.Tasks[m.CurrentTask]
					rt.Create(b)
					return
				}
				switch cj.State {
				case "incomplete":
					b = cj
					code = http.StatusAccepted
					return
				case "finished":
					// Advance to the next task
					if advance {
						m.CurrentTask += 1
					}
				case "failed":
					// Someone has set the machine back to runnable and wants
					// to rerun the current task again.  Let them
				default:
					// Need to error - running job already running or just created.
					err = &models.Error{Code: http.StatusConflict, Type: "Conflict",
						Messages: []string{fmt.Sprintf("Machine %s already has running or created job", b.Machine.String())}}
					code = http.StatusConflict
					return
				}
				if m.CurrentTask >= len(m.Tasks) {
					code = http.StatusNoContent
					return
				}
				// Fill in new job.
				b.State = "created"
				if m.CurrentJob == nil {
					b.Previous = uuid.Parse("00000000-0000-0000-0000-000000000000")
				} else {
					b.Previous = m.CurrentJob
				}
				b.Stage = m.Stage
				b.Task = m.Tasks[m.CurrentTask]
				// Create the job, and then update the machine
				_, err = rt.Create(b)
				if err == nil {
					m.CurrentJob = b.Uuid
					_, err = rt.Save(m)
					if err != nil {
						rt.Remove(b)
						code = http.StatusBadRequest
					}
				}
				code = http.StatusCreated
			})
			switch code {
			case http.StatusAccepted, http.StatusCreated:
				s, ok := models.Model(b).(Sanitizable)
				if ok {
					res = s.Sanitize()
				} else {
					res = b
				}
				c.JSON(code, res)
			case http.StatusNoContent:
				c.Data(http.StatusNoContent, gin.MIMEJSON, nil)
			default:
				if err != nil {
					be, ok := err.(*models.Error)
					if ok {
						c.JSON(be.Code, be)
					} else {
						c.JSON(http.StatusBadRequest, models.NewError(c.Request.Method, http.StatusBadRequest, err.Error()))
					}
				}
			}
		})

	// swagger:route GET /jobs/{uuid} Jobs getJob
	//
	// Get a Job
	//
	// Get the Job specified by {uuid} or return NotFound.
	//
	//     Responses:
	//       200: JobResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/jobs/:uuid",
		func(c *gin.Context) {
			f.Fetch(c, &backend.Job{}, c.Param(`uuid`))
		})

	// swagger:route HEAD /jobs/{uuid} Jobs headJob
	//
	// See if a Job exists
	//
	// Return 200 if the Job specifiec by {uuid} exists, or return NotFound.
	//
	//     Responses:
	//       200: NoContentResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: NoContentResponse
	f.ApiGroup.HEAD("/jobs/:uuid",
		func(c *gin.Context) {
			f.Exists(c, &backend.Job{}, c.Param(`uuid`))
		})

	// swagger:route PATCH /jobs/{uuid} Jobs patchJob
	//
	// Patch a Job
	//
	// Update a Job specified by {uuid} using a RFC6902 Patch structure
	//
	//     Responses:
	//       200: JobResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       406: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.PATCH("/jobs/:uuid",
		func(c *gin.Context) {
			f.Patch(c, &backend.Job{}, c.Param(`uuid`))
		})

	// swagger:route PUT /jobs/{uuid} Jobs putJob
	//
	// Put a Job
	//
	// Update a Job specified by {uuid} using a JSON Job
	//
	//     Responses:
	//       200: JobResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.PUT("/jobs/:uuid",
		func(c *gin.Context) {
			f.Update(c, &backend.Job{}, c.Param(`uuid`))
		})

	// swagger:route DELETE /jobs/{uuid} Jobs deleteJob
	//
	// Delete a Job
	//
	// Delete a Job specified by {uuid}
	//
	//     Responses:
	//       200: JobResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.DELETE("/jobs/:uuid",
		func(c *gin.Context) {
			f.Remove(c, &backend.Job{}, c.Param(`uuid`))
		})

	// swagger:route GET /jobs/{uuid}/actions Jobs getJobActions
	//
	// Get actions for this job
	//
	// Get actions for the Job specified by {uuid} or return NotFound.
	//
	//     Responses:
	//       200: JobActionsResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.GET("/jobs/:uuid/actions",
		func(c *gin.Context) {
			uuid := c.Param(`uuid`)
			j := &backend.Job{}
			var bad bool
			var err error
			rt := f.rt(c, j.Locks("actions")...)
			rt.Do(func(d backend.Stores) {
				var jo models.Model
				if jo = rt.Find("jobs", uuid); jo == nil {
					err = &models.Error{Code: http.StatusNotFound, Type: backend.ValidationError,
						Messages: []string{fmt.Sprintf("Job %s does not exist", uuid)}}
					bad = true
					return
				}
				j = backend.AsJob(jo)
			})
			if bad {
				c.JSON(err.(*models.Error).Code, err)
				return
			}

			if !f.assureAuth(c, "jobs", "actions", j.AuthKey()) {
				return
			}
			actions, err := j.RenderActions(rt)
			if err != nil {
				be, ok := err.(*models.Error)
				if ok {
					c.JSON(be.Code, be)
				} else {
					c.JSON(http.StatusBadRequest, models.NewError(c.Request.Method, http.StatusBadRequest, err.Error()))
				}
			}
			c.JSON(http.StatusOK, actions)

		})

	// swagger:route GET /jobs/{uuid}/log Jobs getJobLog
	//
	// Get the log for this job
	//
	// Get log for the Job specified by {uuid} or return NotFound.
	//
	//     Produces:
	//       application/octet-stream
	//       application/json
	//
	//     Responses:
	//       200: JobLogResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       500: ErrorResponse
	f.ApiGroup.GET("/jobs/:uuid/log",
		func(c *gin.Context) {
			uuid := c.Param(`uuid`)
			j := &backend.Job{}
			var bad bool
			var err *models.Error
			var path string
			rt := f.rt(c, j.Locks("get")...)
			rt.Do(func(d backend.Stores) {
				var jo models.Model
				if jo = rt.Find("jobs", uuid); jo == nil {
					err = &models.Error{Code: http.StatusNotFound, Type: backend.ValidationError,
						Messages: []string{fmt.Sprintf("Job %s does not exist", uuid)}}
					bad = true
					return
				}
				j = backend.AsJob(jo)
				path = j.LogPath(rt)
			})
			if bad {
				c.JSON(err.Code, err)
				return
			}

			if !f.assureAuth(c, "jobs", "log", j.AuthKey()) {
				return
			}

			c.Writer.Header().Set("Content-Type", "application/octet-stream")
			c.File(path)
		})

	// swagger:route PUT /jobs/{uuid}/log Jobs putJobLog
	//
	// Append the string to the end of the job's log.
	//     Consumes:
	//       application/octet-stream
	//
	//     Produces:
	//       application/json
	//
	//     Responses:
	//       204: NoContentResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       415: ErrorResponse
	//       404: ErrorResponse
	//       500: ErrorResponse
	f.ApiGroup.PUT("/jobs/:uuid/log",
		func(c *gin.Context) {
			if c.Request.Body == nil {
				err := &models.Error{Code: http.StatusBadRequest}
				c.JSON(err.Code, err)
				return
			}
			defer c.Request.Body.Close()
			if c.Request.Header.Get(`Content-Type`) != `application/octet-stream` {
				c.JSON(http.StatusUnsupportedMediaType,
					models.NewError("API ERROR", http.StatusUnsupportedMediaType,
						"job log put must have content-type application/octet-stream"))
				return
			}
			uuid := c.Param(`uuid`)
			j := &backend.Job{}
			var bad bool
			var err *models.Error
			rt := f.rt(c, j.Locks("get")...)
			rt.Do(func(d backend.Stores) {
				var jo models.Model
				if jo = d("jobs").Find(uuid); jo == nil {
					err = &models.Error{Code: http.StatusNotFound, Type: backend.ValidationError,
						Messages: []string{fmt.Sprintf("Job %s does not exist", uuid)}}
					bad = true
					return
				}
				j = backend.AsJob(jo)
			})
			if bad {
				c.JSON(err.Code, err)
				return
			}

			if !f.assureAuth(c, "jobs", "log", j.AuthKey()) {
				return
			}

			rt.Do(func(d backend.Stores) {
				var jo models.Model
				if jo = d("jobs").Find(uuid); jo == nil {
					err = &models.Error{Code: http.StatusNotFound, Type: backend.ValidationError,
						Messages: []string{fmt.Sprintf("Job %s does not exist", uuid)}}
					bad = true
					return
				}
				j = backend.AsJob(jo)

				if err := j.Log(rt, c.Request.Body); err != nil {
					err2 := &models.Error{Code: http.StatusInternalServerError, Type: "Server ERROR",
						Messages: []string{err.Error()}}
					c.JSON(err2.Code, err2)
				} else {
					c.Data(http.StatusNoContent, gin.MIMEJSON, nil)
				}
			})
			if bad {
				c.JSON(err.Code, err)
				return
			}
		})

	job := &backend.Job{}
	pActions, pAction, pRun := f.makeActionEndpoints(job.Prefix(), job, "uuid")

	// swagger:route GET /jobs/{uuid}/plugin_actions Jobs getJobActions
	//
	// List job plugin_actions Job
	//
	// List Job plugin_actions for a Job specified by {uuid}
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       200: ActionsResponse
	//       401: NoJobResponse
	//       403: NoJobResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/jobs/:uuid/plugin_actions", pActions)

	// swagger:route GET /jobs/{uuid}/plugin_actions/{cmd} Jobs getJobAction
	//
	// List specific action for a job Job
	//
	// List specific {cmd} action for a Job specified by {uuid}
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       200: ActionResponse
	//       400: ErrorResponse
	//       401: NoJobResponse
	//       403: NoJobResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/jobs/:uuid/plugin_actions/:cmd", pAction)

	// swagger:route POST /jobs/{uuid}/plugin_actions/{cmd} Jobs postJobAction
	//
	// Call an action on the node.
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       400: ErrorResponse
	//       200: ActionPostResponse
	//       401: NoJobResponse
	//       403: NoJobResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	f.ApiGroup.POST("/jobs/:uuid/plugin_actions/:cmd", pRun)
}

package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// Groups returns a list of groups if requested via GET.
// If requested via POST, APIGroups creates a new group and returns a reference to it.
func (as *Server) Groups(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		gs, err := models.GetGroups(ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "No groups found"}, http.StatusNotFound)
			return
		}
		JSONResponse(w, gs, http.StatusOK)
	// POST: Create a new group and return it as JSON
	case r.Method == "POST":
		g := models.Group{}
		// Put the request into a group

		// Check if content is CSV
		// NB We can only upload one file at a time for new group creation, as three seperate POSTs are sent in quick succesion. We can't
		// identify which group the 1+n files should be assigned to. Need to give this more thought, perhaps we can upload multiple files in one POST,
		// or assign a temporary tracking token to link them together.
		var csvmode = false
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			csvmode = true
			targets, groupname, err := util.ParseCSV(r)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
				return
			}
			g.Name = groupname
			g.Targets = targets
		} else {
			// else decode as JSON
			err := json.NewDecoder(r.Body).Decode(&g)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
				return
			}
		}
		_, err := models.GetGroupByName(g.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Group name already in use"}, http.StatusConflict)
			return
		}
		g.ModifiedDate = time.Now().UTC()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		// With CSV we don't return the entire target list, in line with the new pagination server side processing. To maintain backwards API capabiltiy the JSON request
		// will still return the full list.
		if csvmode == true {
			JSONResponse(w, models.GroupSummary{Id: g.Id, Name: g.Name, ModifiedDate: g.ModifiedDate, NumTargets: int64(len(g.Targets))}, http.StatusCreated)
			return
		}
		JSONResponse(w, g, http.StatusCreated)
	}
}

// GroupsSummary returns a summary of the groups owned by the current user.
func (as *Server) GroupsSummary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		gs, err := models.GetGroupSummaries(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, gs, http.StatusOK)
	}
}

// Group returns details about the requested group.
// If the group is not valid, Group returns null.
func (as *Server) Group(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)

	// Paramters passed by DataTables for pagination are handled below
	v := r.URL.Query()
	search := v.Get("search[value]")
	sortcolumn := v.Get("order[0][column]")
	sortdir := v.Get("order[0][dir]")
	sortby := v.Get("columns[" + sortcolumn + "][data]")
	order := sortby + " " + sortdir // e.g "first_name asc"
	start, err := strconv.ParseInt(v.Get("start"), 0, 64)
	if err != nil {
		start = -1 // Default. gorm will ignore with this value.
	}
	length, err := strconv.ParseInt(v.Get("length"), 0, 64)
	if err != nil {
		length = -1 // Default. gorm will ignore with this value.
	}
	draw, err := strconv.ParseInt(v.Get("draw"), 0, 64)
	if err != nil {
		draw = -1 // If the draw value is missing we can assume this is not a DataTable request and return regular API result
	}

	var g models.Group
	if draw == -1 {
		g, err = models.GetGroup(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
			return
		}
	} else {
		// We don't want to fetch the whole set of targets from a group if we're handling a pagination request. This call
		// is just to validate group ownership
		_, err := models.GetGroupSummary(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
			return
		}
	}

	switch {
	case r.Method == "GET":

		// If draw paratmer is -1 return regular API response, otherwise return pagination response
		if draw == -1 {
			JSONResponse(w, g, http.StatusOK)
		} else {
			// Handle pagination for DataTable
			dT, err := models.GetDataTable(id, start, length, search, order)
			if err != nil {
				log.Errorf("error fetching datatable: %v", err)
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
				return
			}
			dT.Draw = draw
			JSONResponse(w, dT, http.StatusOK)
		}

	case r.Method == "DELETE":
		err = models.DeleteGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting group"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Group deleted successfully!"}, http.StatusOK)
	case r.Method == "PUT":
		// Change this to get from URL and uid (don't bother with id in r.Body)
		g = models.Group{}

		//Check if content is CSV
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			targets, _, err := util.ParseCSV(r)
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
				return
			}
			err = models.AddTargetsToGroup(targets, id)
			if err != nil {
				log.Errorf("error add targets to group: %v", err)
				JSONResponse(w, models.Response{Success: false, Message: "Unable to add targets to group!"}, http.StatusBadRequest)
				return
			}
			// With CSV we don't return the entire target list, in line with the new pagination server side processing.
			ng, err := models.GetGroupSummary(id, ctx.Get(r, "user_id").(int64))
			if err != nil {
				JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
				return
			}
			JSONResponse(w, ng, http.StatusCreated)
			return
		}

		// Default JSON
		err = json.NewDecoder(r.Body).Decode(&g)
		if err != nil {
			log.Errorf("error decoding group: %v", err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}

		if g.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "Error: /:id and group_id mismatch"}, http.StatusInternalServerError)
			return
		}
		g.ModifiedDate = time.Now().UTC()
		g.UserId = ctx.Get(r, "user_id").(int64)

		err = models.PutGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, g, http.StatusOK)
	}
}

// GroupSummary returns a summary of the groups owned by the current user.
func (as *Server) GroupSummary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 0, 64)
		g, err := models.GetGroupSummary(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
			return
		}
		JSONResponse(w, g, http.StatusOK)
	}
}

// GroupTarget handles interactions with individual targets
func (as *Server) GroupTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gid, _ := strconv.ParseInt(vars["id"], 0, 64) // group id

	// Ensure the group belongs to the user
	_, err := models.GetGroupSummary(gid, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
		return
	}
	t := models.Target{}
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Unable to decode target"}, http.StatusInternalServerError)
		return
	}
	switch {
	case r.Method == "PUT":
		// Add an individual target to a group
		err = models.AddTargetsToGroup([]models.Target{t}, gid)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Unable to add target to group"}, http.StatusNotFound)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Added target to group"}, http.StatusCreated)
	case r.Method == "DELETE":

		err := models.DeleteTarget(&t, gid, ctx.Get(r, "user_id").(int64)) // We pass the group id to update modified date, and userid to ensure user owner the target
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Deleted target"}, http.StatusCreated)
	}
}

// GroupRename handles renaming of a group (without supplying all the targets, as in Group() PUT above)
func (as *Server) GroupRename(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64) // group id
	g, err := models.GetGroup(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Group not found"}, http.StatusNotFound)
		return
	}

	switch {
	case r.Method == "PUT":
		g = models.Group{}
		err = json.NewDecoder(r.Body).Decode(&g)
		if err != nil {
			log.Errorf("error decoding group: %v", err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		if g.Id != id {
			JSONResponse(w, models.Response{Success: false, Message: "Error: /:id and group_id mismatch"}, http.StatusInternalServerError)
			return
		}
		_, err := models.GetGroupByName(g.Name, ctx.Get(r, "user_id").(int64))
		if err != gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success: false, Message: "Group name already in use"}, http.StatusConflict)
			return
		}
		g.ModifiedDate = time.Now().UTC()
		g.UserId = ctx.Get(r, "user_id").(int64)
		err = models.UpdateGroup(&g)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Group renamed"}, http.StatusCreated)

	}
}

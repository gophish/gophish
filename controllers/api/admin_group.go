package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

type adminGroupRequest struct {
    Name    string              `json:"name"`
    Users   []models.User       `json:"users" gorm:"association_autoupdate:false;association_autocreate:false;many2many:users_admin_groups;"`
}

func (as *Server) AdminGroups(w http.ResponseWriter, r *http.Request) {
    switch {
    case r.Method == "GET":
        admin_groups, err := models.GetAdminGroups()
        if err != nil {
            JSONResponse(w, models.Response{ Success: false, Message: err.Error() }, http.StatusInternalServerError)
            return
        }
        JSONResponse(w, admin_groups, http.StatusOK)
        return
    case r.Method == "POST":
        admin_group_request := &adminGroupRequest{}
        err := json.NewDecoder(r.Body).Decode(admin_group_request)
        if err != nil {
            JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
            return
        }

        admin_group := models.AdminGroup{
            Name: admin_group_request.Name,
            Users: admin_group_request.Users,
        }

        err = models.PutAdminGroup(&admin_group)
        if err != nil {
            JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
            return
        }

        JSONResponse(w, admin_group, http.StatusOK)
        return
    }
}

func (as *Server) AdminGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)

    existingAdminGroup, err := models.GetAdminGroup(id)
    if err != nil {
        JSONResponse(w, models.Response{Success: false, Message: "Administration Group not found"}, http.StatusNotFound)
        return
    }

    switch {
    case r.Method == "GET":
        JSONResponse(w, existingAdminGroup, http.StatusOK)
    case r.Method == "PUT":
        admin_group_request := &adminGroupRequest{}
        err := json.NewDecoder(r.Body).Decode(admin_group_request)
        if err != nil {
            JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
            return
        }
        existingAdminGroup.Name = admin_group_request.Name
        existingAdminGroup.Users = admin_group_request.Users

        err = models.PutAdminGroup(&existingAdminGroup)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}

        JSONResponse(w, existingAdminGroup, http.StatusOK)
    }
}

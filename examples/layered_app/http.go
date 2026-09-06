package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Controller is the transport boundary. Several services implement it, and the
// container hands all of them over as a slice.
type Controller interface {
	Register(mux *http.ServeMux)
}

type UserController struct {
	service *UserService
}

var _ Controller = (*UserController)(nil)

func NewUserController(service *UserService) *UserController {
	return &UserController{service: service}
}

func (c *UserController) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/{id}", c.getUser)
}

func (c *UserController) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	user, ok := c.service.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

type HealthController struct {
	cfg *Config
}

var _ Controller = (*HealthController)(nil)

func NewHealthController(cfg *Config) *HealthController {
	return &HealthController{cfg: cfg}
}

func (c *HealthController) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"app": c.cfg.AppName, "status": "ok"})
	})
}

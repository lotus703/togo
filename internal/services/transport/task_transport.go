package transport

import (
	"encoding/json"
	"github.com/manabie-com/togo/internal/services/usecase"
	"github.com/manabie-com/togo/internal/storages"
	"log"
	"net/http"
)

type Controller struct {
	ToDoUseCase usecase.ToDoUseCase
}

func (c *Controller) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	log.Println(req.Method, req.URL.Path)
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Set("Access-Control-Allow-Headers", "*")
	resp.Header().Set("Access-Control-Allow-Methods", "*")

	if req.Method == http.MethodOptions {
		resp.WriteHeader(http.StatusOK)
		return
	}

	switch req.URL.Path {
	case "/login":
		c.getAuthToken(resp, req)
		return
	case "/tasks":
		var ok bool
		req, ok = c.ToDoUseCase.ValidToken(req)
		if !ok {
			resp.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch req.Method {
		case http.MethodGet:
			c.listTasks(resp, req)
		case http.MethodPost:
			c.addTask(resp, req)
		}
		return
	}
}
func (c *Controller) getAuthToken(resp http.ResponseWriter, req *http.Request) {
	token, _ := c.ToDoUseCase.GetAuthToken(resp, req)
	json.NewEncoder(resp).Encode(map[string]string{
		"data": token,
	})
}

func (c *Controller) listTasks(resp http.ResponseWriter, req *http.Request) {
	listTasks, err := c.ToDoUseCase.ListTasks(resp, req)
	if err != nil{
		json.NewEncoder(resp).Encode(map[string]string{
			"message": err.Error(),
		})
		return
	}
	json.NewEncoder(resp).Encode(map[string][]*storages.Task{
		"data": listTasks})
}
func (c *Controller) addTask(resp http.ResponseWriter, req *http.Request) {
	task, err := c.ToDoUseCase.AddTask(resp, req)
	if err != nil{
		json.NewEncoder(resp).Encode(map[string]string{
			"message": err.Error(),
		})
		return
	}
	json.NewEncoder(resp).Encode(map[string]*storages.Task{
		"data": task,
	})
}

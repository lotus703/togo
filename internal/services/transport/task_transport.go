package transport

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
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
		req, ok = c.validToken(req)
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
func (c *Controller) getAuthToken(resp http.ResponseWriter, req *http.Request) error {
	id := value(req, "user_id")
	if !c.ToDoUseCase.Store.ValidateUser(req.Context(), id, value(req, "password")) {
		resp.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(resp).Encode(map[string]string{
			"error": "incorrect user_id/pwd",
		})
		return nil
	}
	token, err := c.ToDoUseCase.GetAuthToken(id)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(resp).Encode(map[string]string{
			"error": err.Error(),
		})
	}
	resp.Header().Set("Content-Type", "application/json")
	json.NewEncoder(resp).Encode(map[string]string{
		"data": token,
	})
	return nil
}

func (c *Controller) listTasks(resp http.ResponseWriter, req *http.Request) error {

	id, b := userIDFromCtx(req.Context())
	log.Print(b)

	createDate := value(req, "created_date")
	listTasks, err := c.ToDoUseCase.ListTasks(req.Context(), id, createDate)
	resp.Header().Set("Content-Type", "application/json")
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(resp).Encode(map[string]string{
			"error": err.Error(),
		})
		return err
	}
	json.NewEncoder(resp).Encode(map[string][]*storages.Task{
		"data": listTasks})
	return nil
}

func (c *Controller) addTask(resp http.ResponseWriter, req *http.Request) error {
	t := &storages.Task{}
	err := json.NewDecoder(req.Body).Decode(t)
	if err != nil {
		log.Print("Error")
		return err
	}
	defer req.Body.Close()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return err
	}
	userId, _ := userIDFromCtx(req.Context())
	task, err := c.ToDoUseCase.AddTask(userId,req.Context(), t)
	if err != nil{
		json.NewEncoder(resp).Encode(map[string]string{
			"message": err.Error(),
		})
		return err
	}
	resp.Header().Set("Content-Type", "application/json")
	json.NewEncoder(resp).Encode(map[string]*storages.Task{
		"data": task,
	})
	return nil
}
type userAuthKey int8

func (c *Controller) validToken(req *http.Request) (*http.Request, bool) {
	authHeader := req.Header.Get("Authorization")
	token := authHeader[len("Bearer "):]
	claims := make(jwt.MapClaims)
	t, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) {
		return []byte(c.ToDoUseCase.JWTKey), nil
	})
	if err != nil {
		log.Println(err)
		return req, false
	}
	if !t.Valid {
		return req, false
	}
	id, ok := claims["user_id"].(string)
	if !ok {
		return req, false
	}
	req = req.WithContext(context.WithValue(req.Context(), userAuthKey(0), id))
	return req, true
}
func value(req *http.Request, p string) sql.NullString {
	return sql.NullString{
		String: req.FormValue(p),
		Valid:  true,
	}
}
func  userIDFromCtx(ctx context.Context) (string, bool) {
	v := ctx.Value(userAuthKey(0))
	id, ok := v.(string)
	return id, ok
}

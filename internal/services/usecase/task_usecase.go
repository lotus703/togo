package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/manabie-com/togo/internal/storages"
	"github.com/manabie-com/togo/internal/storages/postgres"
	"log"
	"net/http"
	"time"
)

type userAuthKey int8
type ToDoUseCase struct {
	JWTKey string
	Store  *postgres.Sql
}

//helper
func userIDFromCtx(ctx context.Context) (string, bool) {
	v := ctx.Value(userAuthKey(0))
	id, ok := v.(string)
	return id, ok
}
func value(req *http.Request, p string) sql.NullString {
	return sql.NullString{
		String: req.FormValue(p),
		Valid:  true,
	}
}

func (s *ToDoUseCase) GetAuthToken(resp http.ResponseWriter, req *http.Request) (string, error) {
	id := value(req, "user_id")
	if !s.Store.ValidateUser(req.Context(), id, value(req, "password")) {
		resp.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(resp).Encode(map[string]string{
			"error": "incorrect user_id/pwd",
		})
		return "", nil
	}
	resp.Header().Set("Content-Type", "application/json")

	token, err := s.createToken(id.String)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(resp).Encode(map[string]string{
			"error": err.Error(),
		})
		return "", err
	}
	return token, nil
}

func (s *ToDoUseCase) ListTasks(resp http.ResponseWriter, req *http.Request) ([]*storages.Task, error) {
	id, _ := userIDFromCtx(req.Context())
	tasks, err := s.Store.RetrieveTasks(
		req.Context(),
		sql.NullString{
			String: id,
			Valid:  true,
		},
		value(req, "created_date"),
	)
	resp.Header().Set("Content-Type", "application/json")
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(resp).Encode(map[string]string{
			"error": err.Error(),
		})
		return nil, err
	}
	return tasks, nil
}
func (s *ToDoUseCase) AddTask(resp http.ResponseWriter, req *http.Request) (*storages.Task, error) {
	t := &storages.Task{}
	err := json.NewDecoder(req.Body).Decode(t)
	defer req.Body.Close()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}
	now := time.Now()
	userID, _ := userIDFromCtx(req.Context())
	t.ID = uuid.New().String()
	t.UserID = userID
	t.CreatedDate = now.Format("2006-01-02")
	resp.Header().Set("Content-Type", "application/json")
	maxTodo := s.Store.GetMaximumTask(req.Context(), t)
	count := s.Store.CountTask(req.Context(), t)
	if count >= maxTodo {
		resp.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	err = s.Store.AddTask(req.Context(), t)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(resp).Encode(map[string]string{
			"error": err.Error(),
		})
		return nil, err
	}

	return t, nil
}
func (s *ToDoUseCase) createToken(id string) (string, error) {
	atClaims := jwt.MapClaims{}
	atClaims["user_id"] = id
	atClaims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(s.JWTKey))
	if err != nil {
		return "", err
	}
	return token, nil
}
func (s *ToDoUseCase) ValidToken(req *http.Request) (*http.Request, bool) {
	authHeader := req.Header.Get("Authorization")
	token := authHeader[len("Bearer "):]
	claims := make(jwt.MapClaims)
	t, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) {
		return []byte(s.JWTKey), nil
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

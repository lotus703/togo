package usecase

import (
	"context"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/manabie-com/togo/internal/storages"
	"github.com/manabie-com/togo/internal/storages/postgres"
	"time"
)

type userAuthKey int8
type ToDoUseCase struct {
	JWTKey string
	Store  *postgres.Sql
}

//helper
//func (s *ToDoUseCase)UserIDFromCtx(ctx context.Context) (string, bool) {
//	v := ctx.Value(userAuthKey(0))
//	id, ok := v.(string)
//	return id, ok
//}
func (s *ToDoUseCase) GetAuthToken(id sql.NullString) (string, error) {
	token, err := s.createToken(id.String)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *ToDoUseCase) ListTasks(cxt context.Context,id string, createdDate sql.NullString) ([]*storages.Task, error) {
	tasks, err := s.Store.RetrieveTasks(
		cxt,
		sql.NullString{
			String: id,
			Valid:  true,
		},
		createdDate,
	)
	print(id)
	print(createdDate.String)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
func (s *ToDoUseCase) AddTask(userId string, ctx context.Context,t *storages.Task) (*storages.Task, error) {
	//t := &storages.Task{}
	now := time.Now()
	t.ID = uuid.New().String()
	t.UserID = userId
	t.CreatedDate = now.Format("2006-01-02")
	maxTodo := s.Store.GetMaximumTask(ctx, t)
	count := s.Store.CountTask(ctx, t)
	if count >= maxTodo {
		return nil, nil
	}
	err := s.Store.AddTask(ctx, t)
	if err != nil {
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


package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/manabie-com/togo/internal/storages"

	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
)

type Sql struct {
	Db       *sql.DB
	Host     string
	Port     int
	UserName string
	Password string
	DbName   string
}

func (s *Sql) Connect() {
	dataSource := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		s.Host, s.Port, s.UserName, s.Password, s.DbName)

	var err error
	s.Db, err = sql.Open("postgres", dataSource)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if err := s.Db.Ping(); err != nil {
		log.Error(err.Error())
		return
	}

	fmt.Println("Connect database ok")
}

func (s *Sql) Close() {
	s.Db.Close()
}
func (s *Sql) RetrieveTasks(ctx context.Context, userID, createdDate sql.NullString) ([]*storages.Task, error) {
stmt := `SELECT id, content, user_id, created_date FROM tasks WHERE user_id = $1 AND created_date = $2`
rows, err := s.Db.QueryContext(ctx, stmt, userID, createdDate)
if err != nil {
return nil, err
}
defer rows.Close()

var tasks []*storages.Task
for rows.Next() {
t := &storages.Task{}
err := rows.Scan(&t.ID, &t.Content, &t.UserID, &t.CreatedDate)
if err != nil {
return nil, err
}
tasks = append(tasks, t)
}

if err := rows.Err(); err != nil {
return nil, err
}

return tasks, nil
}

// AddTask adds a new task to DB
func (s *Sql) AddTask(ctx context.Context, t *storages.Task) error {
	stmt := `INSERT INTO tasks (id, content, user_id, created_date) VALUES ($1, $2, $3, $4)`
	_, err := s.Db.ExecContext(ctx, stmt, &t.ID, &t.Content, &t.UserID, &t.CreatedDate)
	if err != nil {
		return err
	}

	return nil
}
// AddTask adds a new task to DB
func (s *Sql) GetMaximumTask(ctx context.Context, t *storages.Task) int {
	stmt := `SELECT max_todo FROM users WHERE id=$1`
	row := s.Db.QueryRowContext(ctx, stmt, &t.UserID)
	var maxTodo int
	err := row.Scan(&maxTodo)
	if err != nil{
		return 0
	}

	return maxTodo
}
//count task on day
func (s *Sql) CountTask(ctx context.Context, t *storages.Task) int  {
	stmt := `SELECT COUNT(id) FROM tasks WHERE user_id=$1 AND created_date=$2`
	row := s.Db.QueryRowContext(ctx, stmt, &t.UserID,&t.CreatedDate)
	var count int
	err := row.Scan(&count)
	if err != nil{
		return 0
	}
	return count
}
// ValidateUser returns tasks if match userID AND password
func (s *Sql) ValidateUser(ctx context.Context, userID, pwd sql.NullString) bool {
	stmt := `SELECT id FROM users WHERE id = $1 AND password = $2`
	row := s.Db.QueryRowContext(ctx, stmt, userID, pwd)
	u := &storages.User{}
	err := row.Scan(&u.ID)
	if err != nil {
		return false
	}

	return true
}

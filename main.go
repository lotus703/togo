package main

import (
	"github.com/manabie-com/togo/internal/services/transport"
	"github.com/manabie-com/togo/internal/services/usecase"
	"github.com/manabie-com/togo/internal/storages/postgres"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	store := &postgres.Sql{
		Host: "localhost",
		Port: 5432,
		UserName: "postgres",
		Password: "root",
		DbName: "togo",
	}

	store.Connect()
	defer  store.Close()
	todoUseCase := usecase.ToDoUseCase{
		Store: store,
		JWTKey: "wqGyEBBfPK9w3Lxw",
	}
	http.ListenAndServe(":5050", &transport.Controller{
		ToDoUseCase: todoUseCase,
	})
}

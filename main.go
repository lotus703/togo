package main

import (
	"github.com/manabie-com/togo/internal/storages/postgres"
	"net/http"

	"github.com/manabie-com/togo/internal/services"
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
	http.ListenAndServe(":5050", &services.ToDoService{
		JWTKey: "wqGyEBBfPK9w3Lxw",
		Store: store,
	})
}

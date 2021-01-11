package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
)
var (
	toDoService ToDoService
)
func TestgetAuthToken(t *testing.T){
	req, err := http.NewRequest("GET", "/login?user_id=firstUser&password=example", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(toDoService.getAuthToken)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"data: "xadfadfadeagasd"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
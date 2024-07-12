package main

import (
	"context"
	"database/sql"
	"testing"

	"gohire/proto/gen/api"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3"
)

func setupExistingDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "db/gohire.db") // Path to your existing DB
	if err != nil {
		return nil, err
	}

	// Ensure test user exists
	schema := `
	DELETE FROM users WHERE id = 'test_user';
	INSERT INTO users (id, username, first_name, last_name) VALUES
		('test_user', 'testusername', 'Test', 'User');
	`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestUpdateUser(t *testing.T) {
	db, err := setupExistingDB()
	if err != nil {
		t.Fatalf("Failed to set up test DB: %v", err)
	}
	defer db.Close()

	apiServer := NewAPIServer(db)

	ctx := context.Background()
	id := "test_user"
	firstname := "UpdatedFirstName"
	lastname := "UpdatedLastName"

	// Create a request
	req := &connect.Request[api.UpdateUserRequest]{
		Msg: &api.UpdateUserRequest{
			Id:        id,
			FirstName: firstname,
			LastName:  lastname,
		},
	}

	// Call UpdateUser
	_, err = apiServer.UpdateUser(ctx, req)
	if err != nil {
		t.Errorf("expected no error from UpdateUser, got %v", err)
	}

	// Verify the user was updated
	getReq := &connect.Request[api.UpdateUserRequest]{
		Msg: &api.UpdateUserRequest{
			Id: id,
		},
	}
	resp, err := apiServer.GetUser(ctx, getReq)
	if err != nil {
		t.Errorf("expected no error from GetUser, got %v", err)
	}
	if resp.Msg.User.FirstName != firstname {
		t.Errorf("expected first name to be updated to %s, got %s", firstname, resp.Msg.User.FirstName)
	}
	if resp.Msg.User.LastName != lastname {
		t.Errorf("expected last name to be updated to %s, got %s", lastname, resp.Msg.User.LastName)
	}
}

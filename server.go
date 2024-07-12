package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gohire/proto/gen/api"
	"net/http"
	"os"
	"strings"
	"text/template"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "db/gohire.db")
	if err != nil {
		panic(err)
	}
	templates, err := template.New("").ParseGlob("./web/templates/*.html")
	if err != nil {
		panic(err)
	}

	websvr := &WebServer{
		mux:       http.NewServeMux(),
		api:       NewAPIServer(db),
		templates: templates,
	}

	websvr.HandleFunc("GET /", websvr.UsersIndex)
	websvr.HandleFunc("POST /user/edit", websvr.EditUser)
	websvr.HandleFunc("GET /users/{id}", websvr.EditUserIndex)

	httpServer := &http.Server{
		Addr:    ":3001",
		Handler: websvr.mux,
	}
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	}

	//http://localhost:3001

}

type WebServer struct {
	mux       *http.ServeMux
	api       *APIServer
	templates *template.Template
}

func (s *WebServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *WebServer) UsersIndex(w http.ResponseWriter, r *http.Request) {
	apireq := &connect.Request[api.GetUsersRequest]{
		Msg: &api.GetUsersRequest{},
	}
	users, err := s.api.GetUsers(r.Context(), apireq)
	if err != nil {
		panic(err)
	}
	err = s.templates.ExecuteTemplate(w, "index", users.Msg)
	if err != nil {
		panic(err)
	}
}

func (s *WebServer) EditUserIndex(w http.ResponseWriter, r *http.Request) {
	id := extractUserIdFromURL(r.URL.Path)
	if id == "" {
		fmt.Println("Params not correct!")
		return
	}

	apireq := &connect.Request[api.UpdateUserRequest]{
		Msg: &api.UpdateUserRequest{
			Id: id,
		},
	}
	user, err := s.api.GetUser(r.Context(), apireq)
	if err != nil {
		panic(err)
	}
	err = s.templates.ExecuteTemplate(w, "edit", user.Msg)
	if err != nil {
		panic(err)
	}
}
func (s *WebServer) EditUser(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	firstName := r.FormValue("firstname")
	lastName := r.FormValue("lastname")

	apireq := &connect.Request[api.UpdateUserRequest]{
		Msg: &api.UpdateUserRequest{
			Id:        id,
			FirstName: firstName,
			LastName:  lastName,
		},
	}
	user, err := s.api.UpdateUser(r.Context(), apireq)
	if err != nil {
		panic(err)
	}
	err = s.templates.ExecuteTemplate(w, "edit", user.Msg)
	if err != nil {
		panic(err)
	}
}

func NewAPIServer(db *sql.DB) *APIServer {
	return &APIServer{db: db}
}

type APIServer struct {
	db *sql.DB
}

func (s *APIServer) GetUsers(
	ctx context.Context,
	req *connect.Request[api.GetUsersRequest],
) (*connect.Response[api.GetUsersResponse], error) {
	//log.Println("Request headers: ", req.Header())
	rows, err := s.db.Query("select * from users order by username")
	if err != nil {
		fmt.Println("Error querying data:", err)
		return nil, err
	}
	defer rows.Close()
	users := make([]*api.User, 0)
	for rows.Next() {
		var (
			id, username, firstname, lastname string
			ignore                            interface{}
		)

		if err := rows.Scan(&id, &ignore, &username, &firstname, &lastname, &ignore); err != nil {
			fmt.Println("Error scanning row:", err)
			return nil, err
		}
		u := &api.User{
			Id:        id,
			Username:  username,
			FirstName: firstname,
			LastName:  lastname,
		}
		users = append(users, u)
		// fmt.Printf("ID: %d, Name: %s, Age: %d\n", id, name, age)
	}

	res := connect.NewResponse(&api.GetUsersResponse{
		Users: users,
	})
	return res, nil
}

func (s *APIServer) GetUser(
	ctx context.Context,
	req *connect.Request[api.UpdateUserRequest],
) (*connect.Response[api.UpdateUserResponse], error) {

	id := req.Msg.GetId()
	fmt.Println("User: ", id)
	row := s.db.QueryRow("select * from users where id = $1", id)

	user := &api.User{}
	var ignore interface{}
	if err := row.Scan(&user.Id, &ignore, &user.Username, &user.FirstName, &user.LastName, &ignore); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No user found
		}
		return nil, err
	}

	res := connect.NewResponse(&api.UpdateUserResponse{
		User: user,
	})
	return res, nil
}

func (s *APIServer) UpdateUser(
	ctx context.Context,
	req *connect.Request[api.UpdateUserRequest],
) (*connect.Response[api.UpdateUserResponse], error) {

	id := req.Msg.GetId()
	firstname := req.Msg.GetFirstName()
	lastname := req.Msg.GetLastName()
	fmt.Println("UpdateUser: ", firstname, lastname, id)

	_, err := s.db.Exec("UPDATE users SET first_name = $1, last_name = $2 WHERE id = $3", firstname, lastname, id)
	if err != nil {
		fmt.Println("Error (db): ", err)
		return nil, errors.New("cannot update user")
	}

	row := s.db.QueryRow("select * from users where id = $1", id)
	updatedUser := &api.User{}
	var ignore interface{}
	if err := row.Scan(&updatedUser.Id, &ignore, &updatedUser.Username, &updatedUser.FirstName, &updatedUser.LastName, &ignore); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No user found
		}
		return nil, err
	}

	res := connect.NewResponse(&api.UpdateUserResponse{
		User: updatedUser,
	})
	return res, nil
}

func extractUserIdFromURL(path string) string {
	// Extract the user ID from the URL path
	fmt.Println("Path: ", path)
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return parts[2] // Extract the user ID
	}
	return ""
}

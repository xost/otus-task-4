package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	createUserTpl  = `INSERT INTO users (login, password, name, email) VALUES ($1, $2, $3, $4)`
	getUserTpl     = `SELECT login, name, email FROM users WHERE id=$1`
	getUserListTpl = `SELECT login, name, email FROM users`
	updateUserTpl  = `UPDATE users SET login=$2, name=$3, email=$4  WHERE id=$1`
	deleteUserTpl  = `DELETE FROM users WHERE id=$1`
)

var (
	createUserStmt  *sql.Stmt
	getUserStmt     *sql.Stmt
	getUserListStmt *sql.Stmt
	updateUserStmt  *sql.Stmt
	deleteUserStmt  *sql.Stmt
)

func readConf() *configModel {
	cfg := &configModel{
		dbHost: "localhost",
		dbPort: "5432",
		dbName: "db",
		dbUser: "root",
		dbPass: "password",
		host:   "localhost",
		port:   "8000",
	}
	dbHost := os.Getenv("DBHOST")
	dbPort := os.Getenv("DBPORT")
	dbName := os.Getenv("DBNAME")
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASS")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if dbHost != "" {
		cfg.dbHost = dbHost
	}
	if dbPort != "" {
		cfg.dbPort = dbPort
	}
	if dbName != "" {
		cfg.dbName = dbName
	}
	if dbUser != "" {
		cfg.dbUser = dbUser
	}
	if dbPass != "" {
		cfg.dbPass = dbPass
	}
	if host != "" {
		cfg.host = host
	}
	if port != "" {
		cfg.port = port
	}
	return cfg
}

func makeDBURL(cfg *configModel) (*sql.DB, error) {
	pgConn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.dbHost, cfg.dbPort, cfg.dbUser, cfg.dbPass, cfg.dbName,
	)
	db, err := sql.Open("postgres", pgConn)
	return db, err
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := readConf()
	db, err := makeDBURL(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	if err = db.PingContext(ctx); err != nil {
		log.Fatal("Failed to check db connection:", err)
	}

	mustPrepareStmts(ctx, db)

	http.HandleFunc("/api/v1/users/create/", createUserHandle)
	http.HandleFunc("/api/v1/users/get/", getUserHandle)
	http.HandleFunc("/api/v1/users/update/", updateUserHandle)
	http.HandleFunc("/api/v1/users/delete/", deleteUserHandle)

	bindOn := fmt.Sprintf("%s:%s", cfg.host, cfg.port)
	if err := http.ListenAndServe(bindOn, nil); err != nil {
		log.Printf("Failed to bind on [%s]: %s", bindOn, err)
	}
}

func mustPrepareStmts(ctx context.Context, db *sql.DB) {
	var err error

	createUserStmt, err = db.PrepareContext(ctx, createUserTpl)
	if err != nil {
		panic(err)
	}

	getUserStmt, err = db.PrepareContext(ctx, getUserTpl)
	if err != nil {
		panic(err)
	}

	getUserListStmt, err = db.PrepareContext(ctx, getUserListTpl)
	if err != nil {
		panic(err)
	}

	updateUserStmt, err = db.PrepareContext(ctx, updateUserTpl)
	if err != nil {
		panic(err)
	}

	deleteUserStmt, err = db.PrepareContext(ctx, deleteUserTpl)
	if err != nil {
		panic(err)
	}
}

func createUserHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("Error: got $s method request, but /api/v1/users/create/ supports only POST method\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only POST method is supported"))
		return
	}
	u := &userModel{}
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		log.Println("Failed to parse user data:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse user data"))
		return
	}
	if err := createUser(u); err != nil {
		log.Println("Failed to create new user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create new user"))
		return
	}
	fmt.Printf("user data: %+v", *u)
}

func getUserHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("Error: got $s method request, but /api/v1/users/get/ supports only GET method\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only GET method is supported"))
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Println("Failed to parse request", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse request"))
		return
	}
	if r.Form.Has("id") {
		idVal := r.Form.Get("id")
		if idVal == "" {
			log.Println("Got wrong user id")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Got wrong user id"))
			return
		}
		id, err := strconv.Atoi(idVal)
		if err != nil {
			log.Println("Got wrong user id")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Got wrong user id"))
			return
		}
		u, err := getUser(id)
		if err != nil {
			log.Println("Failed to get user:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get user"))
			return
		}
		data, _ := json.Marshal(u)
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
		return
	}
	ul, err := getUserList()
	if err != nil {
		log.Println("Failed to get user list:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get user list"))
		return
	}
	data, _ := json.Marshal(ul)
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func updateUserHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		log.Printf("Error: got $s method request, but /api/v1/users/update/ supports only PATCH method\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only PATCH method is supported"))
		return
	}
	u := &updateUserModel{}
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		log.Println("Failed to parse update user data:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse update user data"))
		return
	}
	if err := updateUser(u); err != nil {
		log.Println("Failed to update user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to update user"))
		return
	}
	fmt.Printf("user data: %+v", *u)
}

func deleteUserHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		log.Printf("Error: got $s method request, but /api/v1/users/delete/ supports only DELETE method\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only DELETE method is supported"))
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Println("Failed to parse request", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse request"))
		return
	}
	idVal := r.Form.Get("id")
	if idVal == "" {
		log.Println("Got wrong user id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Got wrong user id"))
		return
	}
	id, err := strconv.Atoi(idVal)
	if err != nil {
		log.Println("Got wrong user id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Got wrong user id"))
		return
	}
	if err = deleteUser(id); err != nil {
		log.Println("Failed to delete user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to delete user"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func createUser(u *userModel) error {
	if _, err := createUserStmt.Exec(u.Login, u.Password, u.Name, u.Email); err != nil {
		return err
	}
	return nil
}

func getUser(id int) (*userModel, error) {
	rows, err := getUserStmt.Query(id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, errors.New("there is no user with specified id")
	}
	login := new(string)
	name := new(string)
	email := new(string)

	if err = rows.Scan(login, name, email); err != nil {
		return nil, err
	}
	return &userModel{
		Login: *login,
		Name:  *name,
		Email: *email,
	}, nil
}

func getUserList() ([]userModel, error) {
	rows, err := getUserListStmt.Query()
	if err != nil {
		return nil, err
	}
	login := new(string)
	name := new(string)
	email := new(string)
	ul := make([]userModel, 0)
	for rows.Next() {
		if err = rows.Scan(login, name, email); err != nil {
			continue
		}
		ul = append(ul, userModel{
			Login: *login,
			Name:  *name,
			Email: *email,
		})
	}
	if len(ul) == 0 {
		return nil, errors.New("where is no any users")
	}
	return ul, nil
}

func updateUser(u *updateUserModel) error {
	if _, err := updateUserStmt.Exec(u.ID, u.Login, u.Name, u.Email); err != nil {
		return err
	}
	return nil
}

func deleteUser(id int) error {
	res, err := deleteUserStmt.Exec(id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user does not exist")
	}
	return nil
}

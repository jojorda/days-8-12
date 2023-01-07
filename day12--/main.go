package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"my-project/connection"
	"my-project/middleware"
	"strings"

	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()

	// Connect to Database
	connection.DatabaseConnect()

	// for public folder
	// ex: localhost:port/public/ +../path/to/file
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.HandleFunc("/", newHome).Methods("GET")

	// CRUD Project
	route.HandleFunc("/create-project", createProject).Methods("GET")
	route.HandleFunc("/store-project", middleware.UploadFile(storeProject)).Methods("POST")
	route.HandleFunc("/detail-project/{id}", detailProject).Methods("GET")
	route.HandleFunc("/edit-project/{id}", editProject).Methods("GET")
	route.HandleFunc("/edit-project/{id}", middleware.UploadFile(updateProject)).Methods("POST")
	route.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/register", registerForm).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")
	route.HandleFunc("/login", loginForm).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")
	// Logout
	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server berjalan pada port 5000")
	http.ListenAndServe("localhost:5000", route)
}

type Project struct {
	ID           int
	ProjectName  string
	StartDate    time.Time
	EndDate      time.Time
	Duration     string
	Description  string
	Technologies []string
	Image        string
	UserId       int
}

type MetaData struct {
	Id		  int
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = MetaData{}

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

// newHome
func newHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/index.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	dataProjects, errQuery := connection.Conn.Query(context.Background(), "SELECT id, project_name, description, technologies, image FROM tb_projects")
	if errQuery != nil {
		fmt.Println("Message : " + errQuery.Error())
		return
	}

	var result []Project

	for dataProjects.Next() {
		var each = Project{}

		err := dataProjects.Scan(&each.ID, &each.ProjectName, &each.Description, &each.Technologies, &each.Image)
		if err != nil {
			fmt.Println("Message : " + err.Error())
			return
		}

		result = append(result, each)
	}

	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}
	Data.FlashData = strings.Join(flashes, "")
	// fmt.Println(result)
	listProject := map[string]interface{}{
		"Projects": result,
		"Data":     Data,
	}

	tmpt.Execute(w, listProject)
}

// CRUD Project
//
// Project Struct

// createProject
func createProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/create-project.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	//session
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}
	data := map[string]interface{}{
		"Data": Data,
	}
	tmpt.Execute(w, data)
}

// storeProject
func storeProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	project_name := r.PostForm.Get("project_name")
	technologies := r.Form["technologies"]
	description := r.PostForm.Get("description")

	// Image
	dataContext := r.Context().Value("dataFile")
	image_path := dataContext.(string)
	// Duration
	// Date
	const (
		layoutISO = "2006-01-02"
	)
	tStartDate, _ := time.Parse(layoutISO, r.PostForm.Get("start_date"))
	tEndDate, _ := time.Parse(layoutISO, r.PostForm.Get("end_date"))

	// Session
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}
	// fmt.Println(Data.Id)

	user_id := session.Values["Id"].(int)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_projects(project_name, start_date, end_date, technologies, description, image, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7)", project_name, tStartDate, tEndDate, description, technologies, image_path, user_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	// projects = append(projects, newProject)
	// fmt.Println(projects)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// detailProject
func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/detail-project.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var DataProject = Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_projects WHERE id=$1", id).Scan(
		&DataProject.ID, &DataProject.ProjectName, &DataProject.StartDate, &DataProject.EndDate, &DataProject.Description, &DataProject.Technologies, &DataProject.Image,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}
	EditProject := map[string]interface{}{
		"Project": DataProject,
		"Data":    Data,
	}
	// fmt.Println(EditProject)
	tmpt.Execute(w, EditProject)
}

// editProject
func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/edit-project.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	// session
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var DataProject = Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_projects WHERE id=$1", id).Scan(
		&DataProject.ID, &DataProject.ProjectName, &DataProject.StartDate, &DataProject.EndDate, &DataProject.Description, &DataProject.Technologies, &DataProject.Image,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	EditProject := map[string]interface{}{
		"Project": DataProject,
		"Data":    Data,
	}
	// fmt.Println(EditProject)
	tmpt.Execute(w, EditProject)
}

// updateProject
func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	project_name := r.PostForm.Get("project_name")
	technologies := r.Form["technologies"]
	description := r.PostForm.Get("description")

	// Image
	dataContext := r.Context().Value("dataFile")
	image_path := dataContext.(string)
	// Retrieve the image from form data
	// Date
	const (
		layoutISO = "2006-01-02"
	)
	start_date, _ := time.Parse(layoutISO, r.PostForm.Get("start_date"))
	end_date, _ := time.Parse(layoutISO, r.PostForm.Get("end_date"))
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET project_name = $1, start_date = $2, end_date = $3, technologies = $4, description = $5, image = $6 WHERE id = $7", project_name, start_date, end_date, technologies, description, image_path, id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// deleteProject
func deleteProject(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	// projects = append(projects[:id], projects[id+1:]...)

	http.Redirect(w, r, "/", http.StatusFound)
}

func registerForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/register.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	// Session
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] == true {
		Data.IsLogin = true
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	tmpt.Execute(w, nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")

	password := r.PostForm.Get("password")
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_users(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	session.AddFlash("Successfully registered!", "message")

	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func loginForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/login.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	// Session
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] == true {
		Data.IsLogin = true
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	tmpt.Execute(w, nil)
}

// login - login
func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_users WHERE email=$1", email).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password,
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	// Session
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	session.Values["IsLogin"] = true
	session.Values["Name"] = user.Name
	session.Options.MaxAge = 10800 // 3 hours

	session.AddFlash("Successfully login!", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// Logout
func logout(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")
	session.Options.MaxAge = -1

	session.Save(r, w)

	fmt.Println("Logout")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// contact
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/contact.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	// Session
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}
	data := map[string]interface{}{
		"Data": Data,
	}

	tmpt.Execute(w, data)
}

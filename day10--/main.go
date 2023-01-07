package main

import (
	"context"
	"fmt"
	"html/template"

	// "image"
	"io"
	"log"
	"my-project/connection"

	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
)

func main() {
	route := mux.NewRouter()

	// Connect to Database
	connection.DatabaseConnect()

	// for public folder
	// ex: localhost:port/public/ +../path/to/file
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.HandleFunc("/", Home).Methods("GET")

	// CRUD Project
	// create
	route.HandleFunc("/create-project", createProject).Methods("GET")
	// store
	route.HandleFunc("/store-project", storeProject).Methods("POST")
	// detail
	route.HandleFunc("/detail-project/{id}", detailProject).Methods("GET")
	// edit
	route.HandleFunc("/edit-project/{id}", editProject).Methods("GET")
	route.HandleFunc("/edit-project/{id}", updateProject).Methods("POST")
	// delete
	route.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	// contact
	route.HandleFunc("/contact", contact).Methods("GET")

	fmt.Println("Server berjalan pada port 5000")
	http.ListenAndServe("localhost:5000", route)
}

// newHome
func Home(w http.ResponseWriter, r *http.Request) {
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

	// fmt.Println(result)
	listProject := map[string]interface{}{
		"Projects": result,
	}

	tmpt.Execute(w, listProject)
}

// CRUD Project
//
// Project Struct
type Project struct {
	ID           int
	ProjectName  string
	StartDate    time.Time
	EndDate      time.Time
	Duration     string
	Description  string
	Technologies []string
	Image        string
}

// createProject
func createProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/create-project.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	tmpt.Execute(w, nil)
}

// storeProject
func storeProject(w http.ResponseWriter, r *http.Request) {
	// left shift 32 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	err := r.ParseMultipartForm(32 << 20)

	if err != nil {
		log.Fatal(err)
	}

	project_name := r.PostForm.Get("project_name")
	technologies := r.Form["technologies"]
	description := r.PostForm.Get("description")

	// Duration
	// Date
	const (
		layoutISO = "2006-01-02"
	)
	tStartDate, _ := time.Parse(layoutISO, r.PostForm.Get("start_date"))
	tEndDate, _ := time.Parse(layoutISO, r.PostForm.Get("end_date"))

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_projects(project_name, start_date, end_date, description, technologies, image) VALUES ($1, $2, $3, $4, $5, $6)", project_name, tStartDate, tEndDate, description, technologies)
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

	EditProject := map[string]interface{}{
		"Project": DataProject,
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
	}
	// fmt.Println(EditProject)
	tmpt.Execute(w, EditProject)
}

// updateProject
func updateProject(w http.ResponseWriter, r *http.Request) {

	// left shift 32 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	err := r.ParseMultipartForm(32 << 20)

	if err != nil {
		log.Fatal(err)
	}

	project_name := r.PostForm.Get("project_name")
	technologies := r.Form["technologies"]
	description := r.PostForm.Get("description")

	// Image
	// Retrieve the image from form data
	uploadedFile, handler, err := r.FormFile("image")
	if err != nil {
		w.Write([]byte("Error message upload file: " + err.Error()))
		return
	}
	defer uploadedFile.Close()
	fileLocation := "public/uploads/"
	imageName := slug.Make(project_name)
	_ = os.MkdirAll(fileLocation, os.ModePerm)
	fullPath := fileLocation + imageName + filepath.Ext(handler.Filename)
	targetFile, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		w.Write([]byte("Error message target file: " + err.Error()))
		return
	}
	defer targetFile.Close()
	// Copy the file to the destination path
	_, err = io.Copy(targetFile, uploadedFile)
	if err != nil {
		w.Write([]byte("Error message copy file: " + err.Error()))
		return
	}
	// Image Path
	imagePath := fileLocation + imageName + filepath.Ext(handler.Filename)
	// End For Image

	// Duration
	// Date
	const (
		layoutISO = "2006-01-02"
	)
	tStartDate, _ := time.Parse(layoutISO, r.PostForm.Get("start_date"))
	tEndDate, _ := time.Parse(layoutISO, r.PostForm.Get("end_date"))
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET project_name = $1, start_date = $2, end_date = $3, description = $4, technologies = $5, image = $6 WHERE id = $7", project_name, tStartDate, tEndDate, description, technologies, imagePath, id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	// fmt.Println(projects)

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

// contact
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/contact.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	tmpt.Execute(w, nil)
}

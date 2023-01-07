package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
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

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/create-project", createProject).Methods("GET")
	route.HandleFunc("/store-project", storeProject).Methods("POST")
	route.HandleFunc("/detail-project/{id}", detailProject).Methods("GET")
	route.HandleFunc("/edit-project/{id}", editProject).Methods("GET")
	route.HandleFunc("/update-project/{id}", updateProject).Methods("POST")
	route.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")

	fmt.Println("Server berjalan pada port 5000")
	http.ListenAndServe("localhost:5000", route)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/index.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	listProject := map[string]interface{}{
		"Projects": projects,
	}

	tmpt.Execute(w, listProject)
}

type Project struct {
	ID           int
	ProjectName  string
	StartDate    string
	EndDate      string
	Duration     string
	Description  string
	Technologies []string
	Image        string
}

var projects = []Project{
	{
		ProjectName:  "Samu Example",
		StartDate:    "2022-11-26",
		EndDate:      "2022-11-27",
		Duration:     "5 month",
		Description:  "Lorem ipsum, dolor sit amet consectetur adipisicing elit.",
		Technologies: []string{"nodejs", "vuejs", "reactjs", "nextjs"},
		Image:        "public/img/13.jpg",
	},
}

func createProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/create-project.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	tmpt.Execute(w, nil)
}

func storeProject(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(32 << 20)

	if err != nil {
		log.Fatal(err)// langsng menutup tdk usah di return lgi
	}

	project_name := r.PostForm.Get("project_name")
	technologies := r.Form["technologies"]
	description := r.PostForm.Get("description")

	uploadedFile, handler, err := r.FormFile("image")
	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	defer uploadedFile.Close()
	dataLocation := "public/uploads/"
	imageName := slug.Make(project_name)
	_ = os.MkdirAll(dataLocation, os.ModePerm)
	AllData:= dataLocation + imageName + filepath.Ext(handler.Filename)
	targetFile, err := os.OpenFile(AllData, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, uploadedFile)
	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	imagePath := dataLocation + imageName + filepath.Ext(handler.Filename)
	
	startDate := r.PostForm.Get("start_date")
	endDate := r.PostForm.Get("end_date")
	const (
		AllDate = "2011-05-02"
	)
	tStartDate, _ := time.Parse(AllDate, startDate)
	tEndDate, _ := time.Parse(AllDate, endDate)
	diff := tEndDate.Sub(tStartDate)

	months := int64(diff.Hours() / 24 / 30)
	days := int64(diff.Hours() / 24)

	if days%30 >= 0 {
		days = days % 30
	}

	var duration string

	if months >= 1 && days >= 1 {
		duration = strconv.FormatInt(months, 10) + " month " + strconv.FormatInt(days, 10) + " days"
	} else if months >= 1 && days <= 0 {
		duration = strconv.FormatInt(months, 10) + " month"
	} else if months < 1 && days >= 0 {
		duration = strconv.FormatInt(days, 10) + " days"
	} else {
		duration = "1 days"
	}

	var newProject = Project{
		ProjectName:  project_name,
		StartDate:    startDate,
		EndDate:      endDate,
		Duration:     duration,
		Description:  description,
		Technologies: technologies,
		Image:        imagePath,
	}

	projects = append(projects, newProject)

	http.Redirect(w, r, "/", http.StatusMovedPermanently) //
}

func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8") // untuk ngirimkan ke html
	tmpt, err := template.ParseFiles("views/detail-project.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var DataProject = Project{}

	for index, data := range projects {
		if index == id {
			DataProject = Project{
				ID:           id,
				ProjectName:  data.ProjectName,
				StartDate:    data.StartDate,
				EndDate:      data.EndDate,
				Duration:     data.Duration,
				Description:  data.Description,
				Technologies: data.Technologies,
				Image:        data.Image,
			}
		}
	}

	EditProject := map[string]interface{}{
		"Project": DataProject,
	}
	tmpt.Execute(w, EditProject)
}
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/contact.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	tmpt.Execute(w, nil)
}


func updateProject(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(32 << 20)

	if err != nil {
		log.Fatal(err)
	}

	project_name := r.PostForm.Get("project_name")
	technologies := r.Form["technologies"]
	description := r.PostForm.Get("description")

	uploadedFile, handler, err := r.FormFile("image")
	if err != nil {
		w.Write([]byte("Error message upload file: " + err.Error()))
		return
	}
	defer uploadedFile.Close()
	dataLocation := "public/uploads/"
	imageName := slug.Make(project_name)
	_ = os.MkdirAll(dataLocation, os.ModePerm)
	AllData := dataLocation + imageName + filepath.Ext(handler.Filename)
	targetFile, err := os.OpenFile(AllData, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		w.Write([]byte("Error message target file: " + err.Error()))
		return
	}
	defer targetFile.Close()
	
	_, err = io.Copy(targetFile, uploadedFile)
	if err != nil {
		w.Write([]byte("Error message copy file: " + err.Error()))
		return
	}
	
	imagePath := dataLocation + imageName + filepath.Ext(handler.Filename)
	
	startDate := r.PostForm.Get("start_date")
	endDate := r.PostForm.Get("end_date")
	const (
		AllDate = "2007-11-02"
	)
	tStartDate, _ := time.Parse(AllDate, startDate)
	tEndDate, _ := time.Parse(AllDate, endDate)
	diff := tEndDate.Sub(tStartDate)

	months := int64(diff.Hours() / 24 / 30)
	days := int64(diff.Hours() / 24)

	if days%30 >= 0 {
		days = days % 30
	}

	var duration string

	if months >= 1 && days >= 1 {
		duration = strconv.FormatInt(months, 10) + " month " + strconv.FormatInt(days, 10) + " days"
	} else if months >= 1 && days <= 0 {
		duration = strconv.FormatInt(months, 10) + " month"
	} else if months < 1 && days >= 0 {
		duration = strconv.FormatInt(days, 10) + " days"
	} else {
		duration = "0 days"
	}
	
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	projects[id].ProjectName = project_name
	projects[id].ProjectName = project_name
	projects[id].StartDate = startDate
	projects[id].EndDate = endDate
	projects[id].Duration = duration
	projects[id].Description = description
	projects[id].Technologies = technologies
	projects[id].Image = imagePath

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	projects = append(projects[:id], projects[id+1:]...)

	http.Redirect(w, r, "/", http.StatusFound)
}
func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/edit-project.html")

	if err != nil {
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var DataProject = Project{}

	for index, data := range projects {
		if index == id {
			DataProject = Project{
				ID:           id,
				ProjectName:  data.ProjectName,
				StartDate:    data.StartDate,
				EndDate:      data.EndDate,
				Description:  data.Description,
				Technologies: data.Technologies,
				Image:        data.Image,
			}
		}
	}

	EditProject := map[string]interface{}{
		"Project": DataProject,
	}

	tmpt.Execute(w, EditProject)
}



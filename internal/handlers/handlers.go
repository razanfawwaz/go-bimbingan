package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/razanfawwaz/bimbingan/internal/db" // Add this line to import the package

	"github.com/oklog/ulid/v2"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	tmpl.Execute(w, nil)
}

func AddDataHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)
		fmt.Print(err)
		return
	}

	name := r.FormValue("name")
	npm := r.FormValue("npm")
	fieldInterest := r.FormValue("field_interest")
	projectTitle := r.FormValue("project_title")
	batch := r.FormValue("batch")
	picture := r.FormValue("picture")

	id := ulid.Make().String()

	_, err := db.DB.Exec("INSERT INTO students (id, name, npm, field_interest, project_title, batch, picture) VALUES ($1, $2, $3, $4, $5, $6, $7)", id, name, npm, fieldInterest, projectTitle, batch, picture)
	if err != nil {
		http.Error(w, "cannot insert data", http.StatusInternalServerError)
		fmt.Print(err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/admin.html"))
	tmpl.Execute(w, nil)
}

func GraduatesListHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT name, npm, field_interest, project_title, batch, picture FROM students")
	if err != nil {
		http.Error(w, "cannot get data", http.StatusInternalServerError)
		fmt.Print(err)
		return
	}
	defer rows.Close()

	var students []struct {
		Name          string
		NPM           string
		FieldInterest string
		ProjectTitle  string
		Batch         string
		Picture       string
	}

	for rows.Next() {
		var student struct {
			Name          string
			NPM           string
			FieldInterest string
			ProjectTitle  string
			Batch         string
			Picture       string
		}

		if err := rows.Scan(&student.Name, &student.NPM, &student.FieldInterest, &student.ProjectTitle, &student.Batch, &student.Picture); err != nil {
			http.Error(w, "cannot get data", http.StatusInternalServerError)
			return
		}

		students = append(students, student)
	}

	tmpl := template.Must(template.ParseFiles("./templates/partials/graduates.html"))
	tmpl.Execute(w, students)
}

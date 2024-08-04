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
		return
	}

	name := r.FormValue("name")
	npm := r.FormValue("npm")
	fieldInterest := r.FormValue("field_interest")
	projectTitle := r.FormValue("project_title")
	batch := r.FormValue("batch")
	picture := r.FormValue("picture")

	id := ulid.Make().String()

	if picture == "" {
		picture = "https://cdn-icons-png.flaticon.com/512/149/149071.png"
	}

	_, err := db.DB.Exec("INSERT INTO students (id, name, npm, field_interest, project_title, batch, picture) VALUES ($1, $2, $3, $4, $5, $6, $7)", id, name, npm, fieldInterest, projectTitle, batch, picture)
	if err != nil {
		http.Error(w, "cannot insert data", http.StatusInternalServerError)
		return
	}
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/admin.html"))
	tmpl.Execute(w, nil)
}

func GraduatesListHandler(w http.ResponseWriter, r *http.Request) {
	fieldInterest := r.URL.Query().Get("fieldInterest")
	batch := r.URL.Query().Get("batch")

	query := "SELECT name, npm, field_interest, project_title, batch, picture FROM students WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if fieldInterest != "" {
		query += fmt.Sprintf(" AND field_interest = $%d", argCount)
		args = append(args, fieldInterest)
		argCount++
	}

	if batch != "" {
		query += fmt.Sprintf(" AND batch = $%d", argCount)
		args = append(args, batch)
		argCount++
	}

	// Print the query and args for debugging

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
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
		var s struct {
			Name          string
			NPM           string
			FieldInterest string
			ProjectTitle  string
			Batch         string
			Picture       string
		}
		if err := rows.Scan(&s.Name, &s.NPM, &s.FieldInterest, &s.ProjectTitle, &s.Batch, &s.Picture); err != nil {
			http.Error(w, "Data scan failed", http.StatusInternalServerError)
			return
		}
		students = append(students, s)
	}

	tmpl := template.Must(template.ParseFiles("./templates/partials/graduates.html"))
	tmpl.Execute(w, students)
}

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/login.html"))
	errorMsg := r.URL.Query().Get("error")
	data := struct {
		Error string
	}{
		Error: errorMsg,
	}
	tmpl.Execute(w, data)
}

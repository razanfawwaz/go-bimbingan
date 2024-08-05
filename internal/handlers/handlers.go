package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/razanfawwaz/bimbingan/internal/db" // Add this line to import the package
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	tmpl.Execute(w, nil)
}

func AddDataHandler(w http.ResponseWriter, r *http.Request) {
	err := AddData(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func AddData(w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	npm := r.FormValue("npm")
	fieldInterest := r.FormValue("field_interest")
	projectTitle := r.FormValue("project_title")
	batch := r.FormValue("batch")
	token := r.FormValue("token")
	projectLink := r.FormValue("project_link")
	profileLink := r.FormValue("profile_link")
	graduateStatus := r.FormValue("is_graduated")

	if graduateStatus == "graduated" {
		graduateStatus = "true"
	} else if graduateStatus == "not_graduated" {
		graduateStatus = "false"
	} else {
		http.Error(w, "invalid graduate status", http.StatusBadRequest)
		return errors.New("invalid graduate status")
	}

	if !CheckToken(token) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return errors.New("invalid token")
	}

	file, _, err := r.FormFile("picture")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return err
	}

	var pictureURL string
	if err == nil {
		defer file.Close()
		buf := bytes.NewBuffer(nil)
		if _, err := buf.ReadFrom(file); err != nil {
			http.Error(w, "failed to read file into buffer", http.StatusInternalServerError)
			return err
		}

		fileBytes := buf.Bytes()
		contentType := http.DetectContentType(fileBytes)
		fileExtension := filepath.Ext(contentType)

		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
			http.Error(w, "invalid file type", http.StatusBadRequest)
			return errors.New("invalid file type")
		}

		s3Service, err := NewR2Service()
		if err != nil {
			http.Error(w, "failed to create R2 service", http.StatusInternalServerError)
			return err
		}

		fileKey := ulid.MustNew(ulid.Now(), nil).String() + fileExtension
		err = s3Service.UploadFileToR2(context.TODO(), fileKey, fileBytes, contentType)
		if err != nil {
			http.Error(w, "failed to upload file to R2", http.StatusInternalServerError)
			return err
		}

		pictureURL = fmt.Sprintf("https://storage-bimbingan.uskkuliahku.dev/%s", fileKey)
	} else {
		pictureURL = "https://cdn-icons-png.flaticon.com/512/149/149071.png"
	}

	id := ulid.Make().String()
	_, err = db.DB.Exec("INSERT INTO students (id, name, npm, field_interest, project_title, batch, picture, profile_link, project_link, is_graduated, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", id, name, npm, fieldInterest, projectTitle, batch, pictureURL, profileLink, projectLink, graduateStatus, "pending")
	if err != nil {
		http.Error(w, "cannot insert data", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Data added successfully"}`)

	return nil
}

// create a function that check token to database table token

func CheckToken(token string) bool {
	var expiredAt time.Time
	err := db.DB.QueryRow("SELECT expired_at FROM token WHERE id = $1", token).Scan(&expiredAt)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if expiredAt.Before(time.Now()) {
		return false
	}

	return true
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/admin.html"))
	tmpl.Execute(w, nil)
}

func GraduatesListHandler(w http.ResponseWriter, r *http.Request) {
	fieldInterest := r.URL.Query().Get("fieldInterest")
	batch := r.URL.Query().Get("batch")

	query := "SELECT name, npm, field_interest, project_title, batch, picture, project_link, profile_link, is_graduated FROM students WHERE status = 'approved'"
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
		Name             string
		NPM              string
		FieldInterest    string
		ProjectTitle     string
		Batch            string
		Picture          string
		ProjectLink      string
		ProfileLink      string
		GraduationStatus bool
	}

	for rows.Next() {
		var s struct {
			Name             string
			NPM              string
			FieldInterest    string
			ProjectTitle     string
			Batch            string
			Picture          string
			ProjectLink      string
			ProfileLink      string
			GraduationStatus bool
		}
		if err := rows.Scan(&s.Name, &s.NPM, &s.FieldInterest, &s.ProjectTitle, &s.Batch, &s.Picture, &s.ProjectLink, &s.ProfileLink, &s.GraduationStatus); err != nil {
			http.Error(w, "Data scan failed", http.StatusInternalServerError)
			fmt.Print(err)
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

func AddDataPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/add-data.html"))
	tmpl.Execute(w, nil)
}

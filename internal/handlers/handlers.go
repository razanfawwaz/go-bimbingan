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

	"github.com/go-playground/validator"
	"github.com/oklog/ulid/v2"
	"github.com/razanfawwaz/bimbingan/internal/db" // Add this line to import the package
	"github.com/razanfawwaz/bimbingan/internal/model"
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

func validateFormData(form *model.StudentForm) error {
	validate := validator.New()

	// Validate struct fields
	if err := validate.Struct(form); err != nil {
		return err
	}
	return nil
}

func AddData(w http.ResponseWriter, r *http.Request) error {

	form := &model.StudentForm{
		Name:          r.FormValue("name"),
		NPM:           r.FormValue("npm"),
		FieldInterest: r.FormValue("field_interest"),
		ProjectTitle:  r.FormValue("project_title"),
		Batch:         r.FormValue("batch"),
		Token:         r.FormValue("token"),
		ProjectLink:   r.FormValue("project_link"),
		ProfileLink:   r.FormValue("profile_link"),
		IsGraduated:   r.FormValue("is_graduated"),
	}

	file, _, err := r.FormFile("picture")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return err
	}

	if err := validateFormData(form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if form.IsGraduated == "graduated" {
		form.IsGraduated = "true"
	} else if form.IsGraduated == "not_graduated" {
		form.IsGraduated = "false"
	} else {
		http.Error(w, "invalid graduate status", http.StatusBadRequest)
		return errors.New("invalid graduate status")
	}

	if !CheckToken(form.Token) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return errors.New("invalid token")
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

		fileKey := ulid.Make().String() + fileExtension
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
	_, err = db.DB.Exec("INSERT INTO students (id, name, npm, field_interest, project_title, batch, picture, profile_link, project_link, is_graduated, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", id, form.Name, form.NPM, form.FieldInterest, form.ProjectTitle, form.Batch, pictureURL, form.ProfileLink, form.ProjectLink, form.IsGraduated, "pending")
	if err != nil {
		http.Error(w, "cannot insert data", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Data added successfully"}`)

	return nil
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/admin.html"))
	tmpl.Execute(w, nil)
}

func AdvisorHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/advisor.html"))
	tmpl.Execute(w, nil)
}

func GraduatesListHandler(w http.ResponseWriter, r *http.Request) {
	fieldInterest := r.URL.Query().Get("fieldInterest")
	batch := r.URL.Query().Get("batch")
	isGraduated := r.URL.Query().Get("isGraduated")

	query := "SELECT name, npm, field_interest, project_title, batch, picture, project_link, profile_link, is_graduated FROM students WHERE status = 'approved'"
	args := []interface{}{}
	argCount := 1

	if fieldInterest != "" {
		query += fmt.Sprintf(" AND field_interest = $%d", argCount)
		args = append(args, fieldInterest)
		argCount++
	}

	if isGraduated != "" {
		query += fmt.Sprintf(" AND is_graduated = $%d", argCount)
		args = append(args, isGraduated)
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

func GetStudentsData(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, name, npm, field_interest, project_title, batch, picture, project_link, profile_link, is_graduated FROM students WHERE status = 'pending'"

	rows, err := db.DB.Query(query)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer rows.Close()

	var students []struct {
		Id               string
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
			Id               string
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
		if err := rows.Scan(&s.Id, &s.Name, &s.NPM, &s.FieldInterest, &s.ProjectTitle, &s.Batch, &s.Picture, &s.ProjectLink, &s.ProfileLink, &s.GraduationStatus); err != nil {
			http.Error(w, "Data scan failed", http.StatusInternalServerError)
			fmt.Print(err)
			return
		}
		students = append(students, s)
	}

	tmpl := template.Must(template.ParseFiles("./templates/partials/graduates-table.html"))
	tmpl.Execute(w, students)
}

func UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	form := &model.UpdateStatusRequest{
		Id:     r.FormValue("id"),
		Status: r.FormValue("status"),
	}

	// update status and updated_at where id = form.Id
	_, err := db.DB.Exec("UPDATE students SET status = $1, updated_at = $2 WHERE id = $3", form.Status, time.Now(), form.Id)
	if err != nil {
		http.Error(w, "cannot update data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Data updated successfully"}`)
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

func CreateToken(w http.ResponseWriter, r *http.Request) {
	token := ulid.Make().String()
	expiredAt := time.Now().Add(time.Hour * 72)

	_, err := db.DB.Exec("INSERT INTO token (id, expired_at) VALUES ($1, $2)", token, expiredAt)
	if err != nil {
		http.Error(w, "cannot create token", http.StatusInternalServerError)
		return
	}
}

func GetAllToken(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, expired_at FROM token")
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tokens []struct {
		Id        string
		ExpiredAt time.Time
	}

	for rows.Next() {
		var t struct {
			Id        string
			ExpiredAt time.Time
		}
		if err := rows.Scan(&t.Id, &t.ExpiredAt); err != nil {
			http.Error(w, "Data scan failed", http.StatusInternalServerError)
			fmt.Print(err)
			return
		}
		tokens = append(tokens, t)
	}

	tmpl := template.Must(template.ParseFiles("./templates/partials/tokens-table.html"))
	tmpl.Execute(w, tokens)
}

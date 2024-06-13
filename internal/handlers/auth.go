package handlers

import (
	"html/template"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/razanfawwaz/bimbingan/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func LoginHandler(jwtKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl := template.Must(template.ParseFiles("./templates/login.html"))
			tmpl.Execute(w, nil)
			return
		}

		var creds Credentials
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		creds.Username = r.FormValue("username")
		creds.Password = r.FormValue("password")

		var storedCreds Credentials
		err = db.DB.QueryRow("SELECT username, password FROM users WHERE username=$1", creds.Username).Scan(&storedCreds.Username, &storedCreds.Password)
		if err != nil {
			http.Redirect(w, r, "/login?error=Invalid username or password", http.StatusSeeOther)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password))
		if err != nil {
			// force redirect to login page
			http.Redirect(w, r, "/login?error=Invalid username or password", http.StatusSeeOther)
			return
		}

		expirationTime := time.Now().Add(60 * time.Minute)
		claims := &Claims{
			Username: creds.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})

		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	}
}

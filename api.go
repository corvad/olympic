package cascade

import (
	"fmt"
	"log"
	"net/http"
)

var database *DB
var accountManager *AccountManager
var linkManager *LinkManager
var mux *http.ServeMux

func createAccount(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	_, err := accountManager.CreateAccount(email, password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("CreateAccount error:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	refreshToken, jwtToken, err := accountManager.Login(email, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Authentication error:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok", "refreshToken":"` + refreshToken + `", "jwtToken":"` + jwtToken + `"}`))
}

func createLink(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	accountID, err := checkJWTMiddleware(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("JWT validation error:", err)
		return
	}
	url := r.Form.Get("url")
	shortUrl := r.Form.Get("shortUrl")
	err = linkManager.CreateLink(url, shortUrl, accountID)
	statusCode := http.StatusBadRequest
	if err == ErrShortUrlExists {
		statusCode = http.StatusConflict
	}
	if err != nil {
		w.WriteHeader(statusCode)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("CreateLink error:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok", "url":"` + url + `", "shortUrl":"` + shortUrl + `"}`))
}

func getLink(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	_, err := checkJWTMiddleware(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("JWT validation error:", err)
		return
	}
	shortUrl := r.Form.Get("shortUrl")
	url, err := linkManager.GetLink(shortUrl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("GetLink error:", err)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func refreshJWT(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("Token")
	login, err := accountManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("RefreshJWT validation error:", err)
		return
	}
	newJWT, err := accountManager.GenerateJWT(login.AccountID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("RefreshJWT error:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok", "jwtToken":"` + newJWT + `"}`))
}

func logout(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	refreshToken := r.Form.Get("refreshToken")
	login, err := accountManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Logout error:", err)
		return
	}
	err = accountManager.Logout(login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Logout error:", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func checkJWTMiddleware(r *http.Request) (uint, error) {
	jwtToken := r.Form.Get("jwtToken")
	accountID, err := accountManager.ValidateJWT(jwtToken)
	if err != nil {
		log.Println("JWT validation error:", err)
		return 0, err
	}
	return accountID, nil
}

func logRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address := r.Header.Get("X-Forwarded-For")
		log.Printf("%s %s %s", address, r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}

func registerRoutes() {
	mux.HandleFunc("/api/auth/createAccount", createAccount)
	mux.HandleFunc("/api/auth/login", login)
	mux.HandleFunc("/api/auth/createLink", createLink)
	mux.HandleFunc("/api/getLink", getLink)
	mux.HandleFunc("/api/auth/refreshJWT", refreshJWT)
	mux.HandleFunc("/api/auth/logout", logout)

}

func Init(dbName string, jwtSigningSecret string) {
	var err error
	database, err = OpenDB(dbName)
	if err != nil {
		log.Panicf("failed to connect database: %v", err)
	}
	log.Println("Connected to Database:", dbName)
	accountManager = &AccountManager{db: database.DB, jwtSigningSecret: jwtSigningSecret}
	linkManager = &LinkManager{db: database.DB}
	mux = http.NewServeMux()
	registerRoutes()
}

func Run(port int) {
	log.Printf("Starting server on port %d...", port)
	loggedMux := logRequest(mux)
	err := http.ListenAndServe(":"+fmt.Sprintf("%d", port), loggedMux)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func Shutdown() {
	if database != nil {
		err := database.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}

package cascade

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var database *DB
var accountManager *AccountManager
var linkManager *LinkManager
var mux *http.ServeMux

type AccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LinkRequest struct {
	Url      string `json:"url"`
	ShortUrl string `json:"shortUrl"`
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, fmt.Sprintf("CreateAccount Invalid Method: %s", r.Method))
		return
	}
	var req AccountRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, fmt.Sprintf("CreateAccount JSON error: %v", err))
		return
	}
	if req.Email == "" || req.Password == "" {
		errorResponse(w, http.StatusBadRequest, "CreateAccount missing fields")
		return
	}
	_, err = accountManager.CreateAccount(req.Email, req.Password)
	if err != nil {
		errorResponse(w, http.StatusConflict, fmt.Sprintf("CreateAccount error: %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Login method not allowed:", r.Method)
		return
	}
	var req AccountRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Login JSON error:", err)
		return
	}
	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Login missing fields")
		return
	}
	refreshToken, jwtToken, err := accountManager.Login(req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Authentication error:", err)
		return
	}
	cookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api",
		MaxAge:   3600 * 24 * 7,
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func createLink(w http.ResponseWriter, r *http.Request) {
	//get context accountID
	accountID, ok := r.Context().Value("accountID").(uint)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("CreateLink missing accountID in context")
		return
	}
	var req LinkRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("CreateLink JSON error:", err)
		return
	}
	url := req.Url
	shortUrl := req.ShortUrl
	if url == "" || shortUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("CreateLink missing fields")
		return
	}
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

/*
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
*/

func logout(w http.ResponseWriter, r *http.Request) {
	// Invalidate the refresh token
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Logout missing cookie:", err)
		return
	}
	login, err := accountManager.ValidateRefreshToken(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Logout invalid cookie token:", err)
		return
	}
	err = accountManager.Logout(login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error"}`))
		log.Println("Logout error:", err)
		return
	}
	cookie = &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api",
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Authorization: ", "")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func registerRoutes() {
	mux.HandleFunc("/api/auth/createAccount", createAccount)
	mux.HandleFunc("/api/auth/login", login)
	mux.HandleFunc("/api/createLink", authMiddleware(createLink, http.MethodPost))
	//mux.HandleFunc("/api/getLink", authMiddleware(getLink, http.MethodPost))
	mux.HandleFunc("/api/auth/logout", authMiddleware(logout, http.MethodPost))

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
	loggedMux := loggingMiddleware(mux)
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

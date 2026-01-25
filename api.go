package cascade

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var database *DB
var keystore *KVStore
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

var cookie = &http.Cookie{
	Name:     "refreshToken",
	Value:    "",
	HttpOnly: true,
	Secure:   true,
	SameSite: http.SameSiteStrictMode,
	Path:     "/",
	MaxAge:   3600 * 24 * 7, // 7 days
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
		errorResponse(w, http.StatusMethodNotAllowed, fmt.Sprintf("Login Invalid Method: %s", r.Method))
		return
	}
	var req AccountRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Login JSON error")
	}
	if req.Email == "" || req.Password == "" {
		errorResponse(w, http.StatusBadRequest, "Login missing fields")
		return
	}
	refreshToken, jwtToken, err := accountManager.Login(req.Email, req.Password)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Login authentication error")
		return
	}
	cookie.Value = refreshToken
	http.SetCookie(w, cookie)
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func createLink(w http.ResponseWriter, r *http.Request) {
	//get context accountID
	accountID, ok := r.Context().Value("accountID").(uint)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "CreateLink missing accountID in context")
		return
	}
	var req LinkRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "CreateLink JSON error")
		return
	}
	url := req.Url
	shortUrl := req.ShortUrl
	if url == "" || shortUrl == "" {
		errorResponse(w, http.StatusBadRequest, "CreateLink missing fields")
		return
	}
	err = linkManager.CreateLink(url, shortUrl, accountID)
	statusCode := http.StatusBadRequest
	if err == ErrShortUrlExists {
		statusCode = http.StatusConflict
	}
	if err != nil {
		errorResponse(w, statusCode, "CreateLink error")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok", "url":"` + url + `", "shortUrl":"` + shortUrl + `"}`))
}

func getLink(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("accountID").(uint)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "GetLink missing accountID in context")
		return
	}
	shortUrl := r.URL.Path[len("/getLink/"):]
	url, err := linkManager.GetLink(shortUrl)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "GetLink error")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func logout(w http.ResponseWriter, r *http.Request) {
	// Invalidate the refresh token
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Logout missing token")
		return
	}
	login, err := accountManager.ValidateRefreshToken(cookie.Value)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Logout invalid cookie token")
		return
	}
	err = accountManager.Logout(login)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Logout could not invalidate token")
		return
	}
	cookie.MaxAge = -1 // Delete the cookie
	http.SetCookie(w, cookie)
	w.Header().Set("Authorization", "")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func registerRoutes() {
	mux.HandleFunc("/auth/createAccount", createAccount)
	mux.HandleFunc("/auth/login", login)
	mux.HandleFunc("/createLink", authMiddleware(createLink, http.MethodPost))
	mux.HandleFunc("/getLink/", authMiddleware(getLink, http.MethodGet))
	mux.HandleFunc("/auth/logout", authMiddleware(logout, http.MethodPost))
}

func Init(db DBConnection, kv KVConnection, jwtSigningSecret string) {
	var err error
	database, err = NewDatabase(db)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	/*
		keystore, err = NewKVStore(kv)
		if err != nil {
			log.Fatalf("Failed to initialize keystore: %v", err)
		}
	*/
	keystore = nil // Temporarily disable keystore usage
	accountManager = &AccountManager{db: database, kv: keystore, jwtSigningSecret: jwtSigningSecret}
	linkManager = &LinkManager{db: database, kv: keystore}
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
		database.Close()
	}
	if keystore != nil {
		keystore.Close()
	}
}

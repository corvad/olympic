package cascade

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
)

func authMiddleware(h http.HandlerFunc, t string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmp, _ := httputil.DumpRequest(r, true)
		log.Println(string(tmp))
		if r.Method != t {
			errorResponse(w, http.StatusMethodNotAllowed, "AuthMiddleware Invalid Method: "+r.Method)
			return
		}
		jwt := r.Header.Get("Authorization")
		if jwt == "" {
			reissueJWT(w, r)
		} else {
			jwt = jwt[len("Bearer "):]
			id, err := accountManager.ValidateJWT(jwt)
			if err != nil {
				// try to reissue
				reissueJWT(w, r)
			} else {
				//add account id in context
				*r = *r.WithContext(context.WithValue(r.Context(), "accountID", id))
			}
		}
		h.ServeHTTP(w, r)
	})
}

func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address := r.Header.Get("X-Forwarded-For")
		log.Printf("%s %s %s", address, r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}

func reissueJWT(w http.ResponseWriter, r *http.Request) {
	//check cookie as fallback
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "AuthMiddleware missing token")
	}
	login, err := accountManager.ValidateRefreshToken(cookie.Value)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "AuthMiddleware invalid cookie token")
	}
	jwt, err := accountManager.GenerateJWT(login.AccountID)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "AuthMiddleware could not generate JWT from cookie")
	}
	w.Header().Set("Authorization", "Bearer "+jwt)
	*r = *r.WithContext(context.WithValue(r.Context(), "accountID", login.AccountID))
}

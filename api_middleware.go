package cascade

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func authMiddleware(h http.HandlerFunc, t string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id uint
		var err error
		tmp, _ := httputil.DumpRequest(r, true)
		log.Println(string(tmp))
		if r.Method != t {
			errorResponse(w, http.StatusMethodNotAllowed, "AuthMiddleware Invalid Method: "+r.Method)
			return
		}
		jwt := r.Header.Get("Authorization")
		if jwt == "" {
			c, err := r.Cookie("refreshToken")
			if err != nil {
				errorResponse(w, http.StatusUnauthorized, fmt.Sprintf("AuthMiddleware missing JWT and refresh token: %v", err))
				return
			}
			jwt, id, err = reissueJWT(*c)
			if err != nil {
				errorResponse(w, http.StatusUnauthorized, fmt.Sprintf("AuthMiddleware missing JWT: %v", err))
				return
			}
		} else {
			jwt = jwt[len("Bearer "):]
			id, err = accountManager.ValidateJWT(jwt)
			if err != nil {
				// could be bad or expired token, try reissuing
				c, err := r.Cookie("refreshToken")
				if err != nil {
					errorResponse(w, http.StatusUnauthorized, fmt.Sprintf("AuthMiddleware invalid JWT and missing refresh token: %v", err))
					return
				}
				jwt, id, err = reissueJWT(*c)
				if err != nil {
					errorResponse(w, http.StatusUnauthorized, fmt.Sprintf("AuthMiddleware invalid JWT: %v", err))
					return
				}
				w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
			}
		}
		*r = *r.WithContext(context.WithValue(r.Context(), "accountID", id))
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

func reissueJWT(c http.Cookie) (string, uint, error) {
	login, err := accountManager.ValidateRefreshToken(c.Value)
	if err != nil {
		return "", 0, fmt.Errorf("AuthMiddleware invalid refresh token: %v", err)
	}
	jwt, err := accountManager.GenerateJWT(login.AccountID)
	if err != nil {
		return "", 0, fmt.Errorf("AuthMiddleware could not generate new JWT: %v", err)
	}
	return jwt, login.AccountID, nil
}

package cascade

import (
	"log"
	"net/http"
)

func errorResponse(w http.ResponseWriter, status int, message string) {
	// send a generic error response since we dont want the client to know details
	// errors are mainly handled client side via status codes
	w.WriteHeader(status)
	_, err := w.Write([]byte(`{"status":"error"}`))
	if err != nil {
		log.Println("Error writing response:", err)
	}
	log.Println(message)
}

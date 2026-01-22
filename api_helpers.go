package cascade

import (
	"log"
	"net/http"
)

func errorResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_, err := w.Write([]byte(`{"status":"error"}`))
	if err != nil {
		log.Println("Error writing response:", err)
	}
	log.Println(message)
}

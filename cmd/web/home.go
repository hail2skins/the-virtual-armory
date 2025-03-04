package web

import (
	"log"
	"net/http"

	"github.com/hail2skins/the-virtual-armory/cmd/web/views/home"
)

func HelloWebHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")

	// If it's a GET request or no name is provided, render the index page
	if r.Method == "GET" || name == "" {
		component := home.Index()
		err = component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Error rendering index page: %v", err)
			return
		}
		return
	}

	// Otherwise render the hello response
	component := home.HelloResponse(name)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering hello response: %v", err)
		return
	}
}

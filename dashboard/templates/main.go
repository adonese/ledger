package main

import (
	"fmt"
	"log"
	"net/http"
)

type GlobalState struct {
	Count int
}

var global GlobalState

func getHandler(w http.ResponseWriter, r *http.Request) {
	users := []Users{
		{Name: "ahmed", Role: "admin", Email: "me@adonese.sd"},
		{Name: "ahmed", Role: "admin", Email: "meks@adonese.sd"},
		{Name: "khalid", Role: "admin", Email: "me1@adonese.sd"},
	}
	component := page(global.Count, 0, users)
	component.Render(r.Context(), w)
}

type Users struct {
	Name  string
	Title string
	Email string
	Role  string
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// Update state.
	r.ParseForm()

	// Check to see if the global button was pressed.
	if r.Form.Has("global") {
		global.Count++
	}
	//TODO: Update session.

	// Display the form.
	getHandler(w, r)
}

func main() {
	// Handle POST and GET requests.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postHandler(w, r)
			return
		}
		getHandler(w, r)
	})

	// Start the server.
	fmt.Println("listening on http://localhost:8000")
	if err := http.ListenAndServe("localhost:8000", nil); err != nil {
		log.Printf("error listening: %v", err)
	}
}

package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	lib "quiz/lib"

	"golang.org/x/net/websocket"
)

type templateData struct {
	Port string
	Page string
}

var (
	data templateData
	tmpl *template.Template
)

var port string = "8080"

func main() {
	fs := http.FileServer(http.Dir("templates/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.Handle("/quiz", websocket.Handler(lib.Quiz))
	http.Handle("/leaderboard", websocket.Handler(lib.Leaderboard))
	http.HandleFunc("/", runHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
func runHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Stat("templates/" + r.URL.Path)
	if r.URL.Path == "/" {
		tmpl = template.Must(template.ParseFiles("templates/index.html"))
	} else if !file.IsDir() && err == nil {
		tmpl = template.Must(template.ParseFiles("templates/" + r.URL.Path))
	} else if file.IsDir() && err == nil {
		_, err2 := os.Stat("templates/" + r.URL.Path + "/index.html")
		if err2 == nil {
			tmpl = template.Must(template.ParseFiles("templates/" + r.URL.Path + "/index.html"))
		} else {
			tmpl = template.Must(template.ParseFiles("templates/error.html"))
		}
	} else {
		tmpl = template.Must(template.ParseFiles("templates/error.html"))
	}
	data = templateData{
		Port: port,
		Page: r.URL.Path,
	}
	if err := tmpl.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		log.Printf("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/binary-kitchen/gokitchenmood/lampen"
	"github.com/ant0ine/go-json-rest/rest"
)

var uploadalert string
var filetowrite string

func handler(w http.ResponseWriter, r *http.Request) {
	title := "moodlights"
	p := &lampen.Lampen{}
	err := p.LoadLampValues(title)
	if err != nil {
		log.Printf("Error loading Config File")
		for i := range p.Values {
			p.Values[i] = "000000"

		}
	}
	t, _ := template.ParseFiles("templates/template.html")
	t.Execute(w, p)

}

func savehandler(w http.ResponseWriter, r *http.Request) {
	p := &lampen.Lampen{}
	for i := 0; i < 10; i++ {
		err := p.Parse(r.FormValue("Lampe"+strconv.Itoa(i)), i)
		if err != nil {
			t, _ := template.ParseFiles("templates/error.html")
			t.Execute(w, p)
			return
		}
	}
	err := p.WriteLampValues("moodlights")
	if err != nil {
		log.Fatal(err)
	}
	p.Send()
	t, _ := template.ParseFiles("templates/success.html")
	t.Execute(w, p)
}

func sethandler(w http.ResponseWriter, r *http.Request) {
	p := &lampen.Lampen{}
	color := r.FormValue("color")
	for i, _ := range p.Values {
		p.Values[i] = color
	}
	p.WriteLampValues("moodlights")
	p.Send()
	http.Redirect(w, r, "/", http.StatusFound)

}

func statichandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func randomhandler(w http.ResponseWriter, r *http.Request) {
	p := &lampen.Lampen{}
	p.SetRandom()
	http.Redirect(w, r, "/", http.StatusFound)

}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s File [-f] \n", os.Args[0])
		return
	}
	lampen.Setup()
	lampen.File = false
	filetowrite = os.Args[1]
	if len(os.Args) > 2 {
		if os.Args[2] == "-f" {
			lampen.File = true
		}
	}

	lampen.Port = filetowrite

	lampe := &lampen.Lampen{}
	lampe.SetLampstosavedValues("moodlights")

	rhandler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
		EnableStatusService:      true,
		XPoweredBy:               "phil-api",
	}
	err := rhandler.SetRoutes(
		rest.RouteObjectMethod("GET", "/lamps", lampe, "GetLamps"),
		rest.RouteObjectMethod("POST", "/lamps", lampe, "PostLamps"),
		&rest.Route{"GET", "/.status",
			func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(rhandler.GetStatus())
			},
		},
		//rest.RouteObjectMethod("GET", "/users/:id", &users, "GetUser"),
		//rest.RouteObjectMethod("PUT", "/users/:id", &users, "PutUser"),
		//rest.RouteObjectMethod("DELETE", "/users/:id", &users, "DeleteUser"),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/save", savehandler)
	http.HandleFunc("/set", sethandler)
	http.HandleFunc("/random", randomhandler)
	http.HandleFunc("/static/", statichandler)
	http.HandleFunc("/templates", statichandler)
	http.Handle("/api/", http.StripPrefix("/api", &rhandler))

	http.ListenAndServe(":8080", nil)
}

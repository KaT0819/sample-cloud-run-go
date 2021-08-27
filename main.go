package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// templateData provides template parameters.
type templateData struct {
	Service  string
	Revision string
}

// Variables used to generate the HTML page.
var (
	data templateData
	tmpl *template.Template
)

type Temps struct {
	notemp *template.Template
	index  *template.Template
	hello  *template.Template
	login  *template.Template
}

var cs *sessions.CookieStore = sessions.NewCookieStore([]byte("secret-key-12345"))

// Template for no-template
func notemp() *template.Template {
	src := "<html></html>"
	tmp, _ := template.New("index").Parse(src)
	return tmp
}

// setup template function
func setupTemp() *Temps {
	temps := new(Temps)

	temps.notemp = notemp()

	// set index template.
	index, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		index = temps.notemp
	}
	temps.index = index

	// set hello template
	hello, err := template.ParseFiles("templates/hello.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		hello = temps.notemp
	}
	temps.hello = hello

	// set hello template
	login, err := template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		login = temps.notemp
	}
	temps.login = login

	return temps
}

// index handler
func index(w http.ResponseWriter, r *http.Request, tmp *template.Template) {
	content := struct {
		Title   string
		Message string
	}{
		Title:   "Index",
		Message: "Index Page",
	}

	err := tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

// hello handler
func hello(w http.ResponseWriter, r *http.Request, tmp *template.Template) {

	// パラメータ取得
	name := r.FormValue("name")

	msg := "template message<br>これはサンプルです。" + name

	if r.Method == "POST" {
		pass := r.FormValue("pass")
		msg = fmt.Sprintf("Name is %s Password is %s", name, pass)
	}

	content := struct {
		Title       string
		Message     string
		Name        string
		Flg         bool
		SubMessage1 string
		SubMessage2 string
		Items       []string
	}{
		Title:       "template title",
		Message:     msg,
		Name:        name,
		Flg:         false,
		SubMessage1: "サブタイトル",
		SubMessage2: "てすと",
		Items:       []string{"あいうえお", "かきくけこ", "１２３４５６７８９０", "!\"#$%&'()", "<b><i><u>htmlタグ</u></i></b>"},
	}
	err := tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

// login handler
func login(w http.ResponseWriter, r *http.Request, tmp *template.Template) {

	ses, err := cs.Get(r, "hello-session")
	if err != nil {
		log.Fatal(err)
	}

	msg := "名前とパスワードを入力してください。"

	if r.Method == "POST" {
		ses.Values["login"] = false
		ses.Values["name"] = nil
		// パラメータ取得
		name := r.FormValue("name")
		pass := r.FormValue("pass")

		if name == pass {
			ses.Values["login"] = true
			ses.Values["name"] = name
		}
		ses.Save(r, w)

		msg = fmt.Sprintf("Name is %s Password is %s", name, pass)
	} else {
		ses.Values["login"] = false
		ses.Values["name"] = nil

		ses.Save(r, w)
	}

	flg, _ := ses.Values["login"].(bool)
	name, _ := ses.Values["name"].(string)

	if flg {
		msg = "logined " + name
	}

	content := struct {
		Title   string
		Message string
	}{
		Title:   "Cookie Session",
		Message: msg,
	}

	err = tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Initialize template parameters.
	service := os.Getenv("K_SERVICE")
	if service == "" {
		service = "???"
	}

	revision := os.Getenv("K_REVISION")
	if revision == "" {
		revision = "???"
	}

	// Prepare template for execution.
	tmpl = template.Must(template.ParseFiles("index.html"))
	data = templateData{
		Service:  service,
		Revision: revision,
	}

	// Define HTTP server.
	http.HandleFunc("/", helloRunHandler)

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	temps := setupTemp()

	// index handle
	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		index(w, r, temps.index)
	})
	// hello handle
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		hello(w, r, temps.hello)
	})
	// login handle
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		login(w, r, temps.login)
	})

	// PORT environment variable is provided by Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Print("Hello from Cloud Run! The container started successfully and is listening for HTTP requests on $PORT")
	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// helloRunHandler responds to requests by rendering an HTML page.
func helloRunHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		log.Printf("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

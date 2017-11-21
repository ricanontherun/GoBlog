package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	DirectoryPosts = "posts"
	DirectoryViews = "views"
)

const (
	RoutePrefixView = "/view/"
	RoutePrefixEdit = "/edit/"
	RoutePrefixSave = "/save/"
	RoutePrefixOops = "/oops/"
)

var (
	ErrorPermissionDenied = errors.New("Permission denied")
	ErrorResourceMissing  = errors.New("Resource missing")
)

// Global which represents the pre-parsed html templates.
var templates *template.Template

type Post struct {
	Title string
	Body  []byte // Makes ioutil operations easier.
}

func getPostPathFromTitle(title string) string {
	return DirectoryPosts + "/" + title + ".txt"
}

func getViewPath(view string) string {
	return DirectoryViews + "/" + view
}

func renderTemplate(writer http.ResponseWriter, view string, data interface{}) error {
	// Templates are the pre parsed views.
	return templates.ExecuteTemplate(writer, getViewPath(view), data)
}

func (page *Post) Save() error {
	if _, err := os.Stat(DirectoryPosts); os.IsNotExist(err) {
		log.Println(err)
		return ErrorResourceMissing
	}

	err := ioutil.WriteFile(getPostPathFromTitle(page.Title), page.Body, 777)

	if err != nil {
		log.Println(err)
		return ErrorPermissionDenied
	}

	return nil
}

func Load(title string) (*Post, error) {
	body, err := ioutil.ReadFile(getPostPathFromTitle(title))

	if err != nil {
		return nil, err // Ideally, we would return our own error.
	}

	return &Post{Title: title, Body: body}, nil
}

func viewPostHandler(writer http.ResponseWriter, request *http.Request) {
	urlTitle := request.URL.Path[len(RoutePrefixView):]

	if len(urlTitle) == 0 {
		fmt.Fprint(writer, "Dude, try entering a blog post title")
		return
	}

	post, err := Load(urlTitle)

	if err != nil {
		http.Redirect(writer, request, RoutePrefixEdit+urlTitle, http.StatusFound)
		return
	}

	renderTemplate(writer, "view.html", post)
}

func editPostHandler(writer http.ResponseWriter, request *http.Request) {
	urlTitle := request.URL.Path[len(RoutePrefixEdit):]

	if len(urlTitle) == 0 {
		fmt.Fprint(writer, "Error: Missing post title")
		return
	}

	post, err := Load(urlTitle)

	if err != nil {
		post = &Post{
			Title: urlTitle,
		}
	}

	renderTemplate(writer, "edit.html", post)
}

func savePostHandler(writer http.ResponseWriter, request *http.Request) {
	urlTitle := request.URL.Path[len(RoutePrefixSave):]

	if len(urlTitle) == 0 {
		fmt.Fprint(writer, "Failed to save post, missing title.")
		return
	}

	post := &Post{
		Title: urlTitle,
		Body:  []byte(request.FormValue("body")),
	}

	err := post.Save()

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(writer, request, RoutePrefixView+urlTitle, http.StatusFound)
}

func errorHandler(writer http.ResponseWriter, request *http.Request) {
	renderTemplate(writer, "500.html", nil)
}

func setupRoutes() {
	http.HandleFunc(RoutePrefixView, viewPostHandler)
	http.HandleFunc(RoutePrefixEdit, editPostHandler)
	http.HandleFunc(RoutePrefixSave, savePostHandler)
	http.HandleFunc(RoutePrefixOops, errorHandler)
}

func main() {
	files, err := ioutil.ReadDir(DirectoryViews)

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}

	return
	// TODO: Check for posts directory here, create if doesn't exist.

	setupRoutes()

	// Parse our views, panicing if an error occurs.
	templates = template.Must(template.ParseFiles("500.html", "edit.html", "view.html"))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

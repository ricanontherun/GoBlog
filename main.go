package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
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

// template.Must is a covenience which says "If I receive a non-nil err, I panic(). Else I return the first argument, a template pointer."
// This will not work when views with identical names are nested in different directories.
var templates = template.Must(template.ParseGlob(DirectoryViews + "/*.html"))
var matchedRoutes = make(map[string]*regexp.Regexp)

type Post struct {
	Title string
	Body  []byte // Makes ioutil operations easier.
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

func getPostPathFromTitle(title string) string {
	return DirectoryPosts + "/" + title + ".txt"
}

func getViewPath(view string) string {
	return DirectoryViews + "/" + view
}

func renderTemplate(writer http.ResponseWriter, view string, data interface{}) error {
	return templates.ExecuteTemplate(writer, view, data)
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

func registerRoute(route string, pattern string, handler http.HandlerFunc) {
	regex, exists := matchedRoutes[pattern]

	if !exists {
		regex = regexp.MustCompile(pattern)
		matchedRoutes[pattern] = regex
	}

	http.HandleFunc(route, func(writer http.ResponseWriter, request *http.Request) {
		if m := regex.FindStringSubmatch(request.URL.Path); m != nil {
			http.NotFound(writer, request)
			return
		}

		handler(writer, request)
	})
}

func setupRoutes() {
	registerRoute(RoutePrefixView, "a", viewPostHandler)
	registerRoute(RoutePrefixEdit, "b", editPostHandler)
	registerRoute(RoutePrefixSave, "b", savePostHandler)
	registerRoute(RoutePrefixOops, "a", errorHandler)
}

func main() {
	setupRoutes()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
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
	template, err := template.ParseFiles(getViewPath(view))

	if err != nil {
		return err
	}

	template.Execute(writer, data)

	return nil
}

func (page *Post) Save() error {
	return ioutil.WriteFile(getPostPathFromTitle(page.Title), page.Body, 0600)
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

	// else carry on as normal.
	post, err := Load(urlTitle)

	if err != nil {
		fmt.Println("post not found, redirecting.")
		// Rather than display an error, redirect them to the edit page.
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
		fmt.Println("redirecting")
		http.Redirect(writer, request, "/view/hello/", http.StatusInternalServerError)
	}

	fmt.Println("Saved! Redirecting to the view page.")
	http.Redirect(writer, request, RoutePrefixView+urlTitle, http.StatusOK)
}

func errorHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("In the error handler")
	renderTemplate(writer, "500.html", nil)
}

func main() {
	http.HandleFunc(RoutePrefixView, viewPostHandler)
	http.HandleFunc(RoutePrefixEdit, editPostHandler)
	http.HandleFunc(RoutePrefixSave, savePostHandler)
	http.HandleFunc(RoutePrefixOops, errorHandler)
	http.ListenAndServe(":8080", nil)
}

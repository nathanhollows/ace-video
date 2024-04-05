package handlers

import (
	"net/http"

	"github.com/nathanhollows/ace-video/flash"
	"github.com/nathanhollows/ace-video/models"
)

// Shows the setup page
func adminSetupHandler(w http.ResponseWriter, r *http.Request) {
	setDefaultHeaders(w)
	data := templateData(r)
	data["title"] = "setup"

	res, err := models.CheckAnyUsers()
	if err != nil {
		flash.Message{
			Title:   "Error",
			Message: err.Error(),
			Style:   flash.Error,
		}.Save(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else if res {
		flash.Message{
			Message: "You have already set up the system",
			Style:   flash.Info,
		}.Save(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data["messages"] = flash.Get(w, r)
	render(w, data, false, "setup")
}

// Handles the setup form submission and creates the first user
func adminSetupPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// Check if any users exist
	res, err := models.CheckAnyUsers()
	if err != nil {
		flash.Message{
			Title:   "Error",
			Message: err.Error(),
			Style:   flash.Error,
		}.Save(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else if res {
		flash.Message{
			Message: "You have already set up the system",
			Style:   flash.Info,
		}.Save(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Create the user
	user := models.NewUser(email, password)
	err = user.Save()
	if err != nil {
		flash.Message{
			Title:   "Error",
			Message: err.Error(),
			Style:   flash.Error,
		}.Save(w, r)
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	flash.Message{
		Title:   "Success",
		Message: "User created successfully",
		Style:   flash.Success,
	}.Save(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

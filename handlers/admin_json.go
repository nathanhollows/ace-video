package handlers

import (
	"net/http"

	"github.com/nathanhollows/ace-video/flash"
)

// adminJ
func adminJSONhandler(w http.ResponseWriter, r *http.Request) {
	setDefaultHeaders(w)
	data := templateData(r)
	data["title"] = "Media"

	data["messages"] = flash.Get(w, r)

	// Render the template
	render(w, data, true, "json_index")
}

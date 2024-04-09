package handlers

import (
	"net/http"

	"github.com/nathanhollows/ace-video/flash"
	"github.com/nathanhollows/ace-video/models"
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

func adminJSONPreviewHandler(w http.ResponseWriter, r *http.Request) {
	setDefaultHeaders(w)
	data := templateData(r)
	data["title"] = "Preview"

	options, err := parseJSONOptions(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	media, err := models.FindMatchingMedia(r.Context(), options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	exportMedia := make(models.Library, len(media))
	copy(exportMedia, media)
	jsonMedia, err := exportMedia.MarshalJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data["media"] = media
	data["json"] = string(jsonMedia)
	data["layout"] = "htmx"

	render(w, data, true, "json_preview")
}

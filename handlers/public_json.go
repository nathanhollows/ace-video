package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nathanhollows/ace-video/models"
)

func publicDataJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "cache, must-revalidate")
	w.Header().Set("Expires", "3600")
	w.Header().Set("Pragma", "public")

	options, err := parseJSONOptions(r)

	media, err := models.FindMatchingMedia(r.Context(), options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(media)
}

// parseJSONOptions parses the options from the url query string
func parseJSONOptions(r *http.Request) (models.JSONOptions, error) {
	options := models.JSONOptions{
		Limit:  10,
		Offset: 0,
		Sort:   "created_at",
		Order:  "desc",
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			return options, err
		}
		options.Limit = limitInt
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			return options, err
		}
		options.Offset = offsetInt
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		options.Sort = sort
	}
	if order := r.URL.Query().Get("order"); order != "" {
		options.Order = order
	}
	if search := r.URL.Query().Get("search"); search != "" {
		options.Search = search
	}
	if tags := r.URL.Query().Get("tags"); tags != "" {
		options.Tags = tags
	}
	if mime := r.URL.Query().Get("type"); mime != "" {
		options.Type = mime
	}
	if id := r.URL.Query().Get("id"); id != "" {
		options.ID = id
	}
	return options, nil
}

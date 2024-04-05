package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/nathanhollows/ace-video/flash"
	"github.com/nathanhollows/ace-video/models"
)

// adminMediaHandler shows the media files
func adminMediaHandler(w http.ResponseWriter, r *http.Request) {
	setDefaultHeaders(w)
	data := templateData(r)
	data["title"] = "Media"

	data["messages"] = flash.Get(w, r)

	media, err := models.FindAllMedia(r.Context())
	if err != nil {
		flash.Message{
			Title:   "Error",
			Message: "Error finding media: " + err.Error(),
			Style:   flash.Error,
		}.Save(w, r)
	} else {
		data["media"] = media
	}

	// Render the template
	render(w, data, true, "media_index")
}

// adminMediaUpdateHandler handles the media update form submission
func adminMediaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	setDefaultHeaders(w)

	id := chi.URLParam(r, "uuid")

	// Get the media
	media, err := models.FindMediaByID(r.Context(), id)
	if err != nil {
		flash.Message{
			Title:   "Error",
			Message: "Error finding media: " + err.Error(),
			Style:   flash.Error,
		}.Save(w, r)
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
		return
	}

	media.Title = r.FormValue("title")
	media.Description = r.FormValue("description")
	media.Caption = r.FormValue("caption")

	err = media.Save(r.Context())
	if err != nil {
		flash.Message{
			Title:   "Error",
			Message: "Error updating media: " + err.Error(),
			Style:   flash.Error,
		}.Save(w, r)
	} else {
		flash.Message{
			Title:   "Success",
			Message: "Media updated successfully",
			Style:   flash.Success,
		}.Save(w, r)
	}

	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
}

// adminMediaUploadHandler handles the media upload form submission
func adminMediaUploadHandler(w http.ResponseWriter, r *http.Request) {
	setDefaultHeaders(w)
	if r.Method != http.MethodPost {
		flash.Message{
			Title:   "Error",
			Message: "Method Not Allowed",
			Style:   flash.Error,
		}.Save(w, r)
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["file"]

	for _, fileHeader := range files {

		// Check the file size
		var maxUploadSize int64 = 1024 * 1024 * 1024 // 1GB
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			flash.Message{
				Title:   "Error",
				Message: "File is too large. There is a 1GB limit.",
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			flash.Message{
				Title:   "Error",
				Message: "Error uploading file: " + err.Error(),
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		defer file.Close()

		// Check the file type
		// Only allow images and videos
		regex, err := regexp.Compile("image/.*|video/.*")
		if err != nil {
			flash.Message{
				Title:   "Error",
				Message: "Error checking file type: " + err.Error(),
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			flash.Message{
				Title:   "Error",
				Message: "Error reading file: " + err.Error(),
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		filetype := http.DetectContentType(buff)
		if !regex.MatchString(filetype) {
			flash.Message{
				Title:   "Error",
				Message: "Invalid file type. Only images and videos are allowed.",
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			flash.Message{
				Title:   "Error",
				Message: "Error reading file: " + err.Error(),
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		// Create the uploads directory if it doesn't exist
		err = os.MkdirAll("assets/media", 0755)
		if err != nil {
			flash.Message{
				Title:   "Error",
				Message: "Error creating uploads directory: " + err.Error(),
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		// Generate a new file name with a UUID
		id := uuid.New().String()
		extension := filepath.Ext(fileHeader.Filename)
		newFileName := id + extension

		// Create the file
		newFile, err := os.Create("assets/media/" + newFileName)
		if err != nil {
			flash.Message{
				Title:   "Error",
				Message: "Error creating file: " + err.Error(),
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		defer newFile.Close()

		// Copy the file
		_, err = io.Copy(newFile, file)
		if err != nil {
			flash.Message{
				Title:   "Error",
				Message: "Error copying file: " + err.Error(),
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			return
		}

		// Save the media to the database
		media := &models.Media{
			ID:       id,
			Title:    fileHeader.Filename,
			FileName: fileHeader.Filename,
			MimeType: fileHeader.Header.Get("Content-Type"),
			FilePath: "/media/" + newFileName,
		}
		err = media.Save(r.Context())
		if err != nil {
			flash.Message{
				Title:   "Error",
				Message: "Error saving media: " + err.Error(),
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
			// Remove the file
			os.Remove("assets/media/" + newFileName)
			return
		}
	}

	flash.Message{
		Title:   "Success",
		Message: "Media uploaded successfully",
		Style:   flash.Success,
	}.Save(w, r)
	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)

}

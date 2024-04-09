package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gomarkdown/markdown"
	"github.com/nathanhollows/ace-video/filesystem"
	"github.com/nathanhollows/ace-video/flash"
	"github.com/nathanhollows/ace-video/models"
	"github.com/nathanhollows/ace-video/sessions"
)

var router *chi.Mux
var server *http.Server

type userContextKey string

func Start() {

	createRoutes()

	server = &http.Server{
		Addr:    os.Getenv("SERVER_ADDR"),
		Handler: router,
	}
	fmt.Println(server.ListenAndServe())
}

func createRoutes() {
	router = chi.NewRouter()
	router.Use(middleware.Compress(5))
	router.Use(middleware.CleanPath)
	router.Use(middleware.StripSlashes)
	router.Use(middleware.RedirectSlashes)

	router.Get("/", adminMediaHandler)

	router.Get("/data.json", publicDataJSONHandler)

	// Session routes
	router.Get("/login", adminLoginHandler)
	router.Post("/login", adminLoginPostHandler)

	// Setup
	router.Get("/setup", adminSetupHandler)
	router.Post("/setup", adminSetupPostHandler)

	router.Route("/admin", func(r chi.Router) {
		r.Use(adminAuthMiddleware)
		r.Get("/json", adminJSONhandler)
		r.Get("/json/preview", adminJSONPreviewHandler)
		r.Get("/", adminMediaHandler)
		r.Route("/media", func(r chi.Router) {
			r.Get("/", adminMediaHandler)
			r.Post("/", adminMediaUploadHandler)
			r.Post("/{uuid}", adminMediaUpdateHandler)
		})
		r.Route("/cards", func(r chi.Router) {
		})
	})

	workDir, _ := os.Getwd()
	filesDir := filesystem.Myfs{Dir: http.Dir(filepath.Join(workDir, "assets"))}
	filesystem.FileServer(router, "/assets", filesDir)

}

func adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the session
		session, err := sessions.Get(r, "admin")
		if err != nil {
			// Destroy the session
			session.Options.MaxAge = -1
			session.Save(r, w)

			flash.Message{
				Title:   "Error",
				Message: "An error occurred while trying to log in.",
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if session.Values["user_id"] == nil {
			flash.Message{
				Title:   "Error",
				Message: "You must be logged in to access this page",
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Find the user by the session
		user, err := models.FindUserBySession(r)
		if err != nil {
			// Destroy the session
			session.Options.MaxAge = -1
			session.Save(r, w)
			flash.Message{
				Title:   "Error",
				Message: "An error occurred while trying to log in.",
				Style:   flash.Error,
			}.Save(w, r)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		key := userContextKey("user")
		ctx := context.WithValue(r.Context(), key, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func templateData(r *http.Request) map[string]interface{} {
	ctxKey := userContextKey("user")
	user, ok := r.Context().Value(userContextKey(ctxKey)).(*models.User)
	data := map[string]interface{}{
		"hxrequest": r.Header.Get("HX-Request") == "true",
		"layout":    "base",
	}
	if ok {
		data["user"] = user
	}
	return data
}

func render(w http.ResponseWriter, data map[string]interface{}, admin bool, patterns ...string) error {
	w.Header().Set("Content-Type", "text/html")

	baseDir := "templates/public/"
	if admin {
		baseDir = "templates/admin/"
	}

	err := parse(data, baseDir, patterns...).ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), 0)
		log.Print("Template executing error: ", err)
	}
	return err
}

func parse(data map[string]interface{}, baseDir string, patterns ...string) *template.Template {
	// Format the title to include the app name
	if title, ok := data["title"].(string); ok {
		data["title"] = fmt.Sprintf("%s | %s", title, os.Getenv("APP_NAME"))
	}

	// Prepend the base directory to each pattern.
	for i, pattern := range patterns {
		patterns[i] = filepath.Join(baseDir, "pages", pattern+".html")
	}

	// Add the components dir to the patterns
	components, err := filepath.Glob(filepath.Join(baseDir, "components", "*.html"))
	if err != nil {
		log.Print("Error getting components: ", err)
	}
	patterns = append(patterns, components...)

	// Get the chosen layout
	if layout, ok := data["layout"].(string); ok {
		patterns = append(patterns, filepath.Join(baseDir, "layouts", layout+".html"))
	}

	// Create a new template, add any functions, and parse the files.
	return template.Must(template.New("base").Funcs(funcs).ParseFiles(patterns...))
}

var funcs = template.FuncMap{
	"html": func(v string) template.HTML {
		return template.HTML(v)
	},
	"upper": func(v string) string {
		return strings.ToUpper(v)
	},
	"lower": func(v string) string {
		return strings.ToLower(v)
	},
	"date": func(t time.Time) string {
		if t.Year() == time.Now().Year() {
			return t.Format("2 January")
		}
		return t.Format("2 January 2006")
	},
	"time": func(t time.Time) string {
		return t.Format("15:04")
	},
	"divide": func(a, b int) float32 {
		if a == 0 || b == 0 {
			return 0
		}
		return float32(a) / float32(b)
	},
	"nl2br": func(s string) template.HTML {
		// Replace newlines with <br> tags
		return template.HTML(strings.Replace(s, "\n", "<br>", -1))
	},
	"progress": func(a, b int) float32 {
		if a == 0 || b == 0 {
			return 0
		}
		return float32(a) / float32(b) * 100
	},
	"add": func(a, b int) int {
		return a + b
	},
	"year": func() string {
		return time.Now().Format("2006")
	},
	"static": func(filename string) string {
		filename = strings.TrimPrefix(filename, "/")
		// get last modified time
		file, err := os.Stat("assets/" + filename)

		if err != nil {
			return "/assets/" + filename
		}

		modifiedtime := file.ModTime()
		return "/assets/" + filename + "?v=" + modifiedtime.Format("20060102150405")
	},
	// Convert a float to a duration and present it in a human readable format
	"toDuration": func(seconds float64) string {
		return time.Duration(int(seconds) * int(time.Second)).String()

	},
	"md": func(s string) template.HTML {
		// Convert markdown to HTML
		content := []byte(s)
		return template.HTML(markdown.ToHTML(content, nil, nil))
	},
}

func setDefaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

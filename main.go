//go:generate npm run build

package main

import (
	"github.com/joho/godotenv"
	"github.com/nathanhollows/ace-video/handlers"
	"github.com/nathanhollows/ace-video/models"
	"github.com/nathanhollows/ace-video/sessions"
)

func main() {
	godotenv.Load(".env")
	models.InitDB()
	sessions.Start()
	handlers.Start()
}

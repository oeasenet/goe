package http

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/gofiber/fiber/v3"
)

//go:embed error_page.gohtml
var templateFS embed.FS

// tpl is a package-level variable that holds the parsed template.
// Parsing it once at startup is efficient.
var tpl *template.Template

// ErrorPageData defines the structure of data passed to the template.
// Fields must be exported (start with a capital letter) to be accessible in the template.
type ErrorPageData struct {
	Title    string
	Code     string
	Message  string
	HomeLink string
}

func ErrorPage(ctx fiber.Ctx, title string, code string, message string, homeLink string) error {
	data := &ErrorPageData{
		Title:    title,
		Code:     code,
		Message:  message,
		HomeLink: homeLink,
	}
	// Create a buffer to temporarily write the template execution result to.
	var buf bytes.Buffer
	err := tpl.Execute(&buf, data)
	if err != nil {
		return ctx.SendString("INTERNAL RENDERING ERROR")
	}
	return ctx.Send(buf.Bytes())
}

package http

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v3"
)

func (m *Http) GoeFiberErrorHandler(ctx fiber.Ctx, err error) error {
	// Status code defaults to 500
	respCode := fiber.StatusInternalServerError
	// Set error message
	message := err.Error()
	// Check if it's a fiber.Error type
	var e *fiber.Error
	if errors.As(err, &e) {
		respCode = e.Code
		message = e.Message
	}
	ctx.Status(respCode)

	// If the format is forced to json or text through query parameter, then return the response in that format
	if ctx.Query("format") == "json" {
		return ctx.JSON(fiber.Map{
			"message": message,
		})
	}

	// If the format is forced to text through query parameter, then return the response in that format
	if ctx.Query("format") == "text" {
		ctx.Response().Header.SetContentType(fiber.MIMETextPlain)
		return ctx.SendString(message)
	}

	if ctx.Accepts(fiber.MIMETextHTML) == fiber.MIMETextHTML {
		// default response, html error page
		ctx.Response().Header.SetContentType(fiber.MIMETextHTML)
		return ctx.SendString(ErrorPage(fmt.Sprintf("ERROR %d", respCode), fmt.Sprintf("%d", respCode), message, "/"))
	}

	// If the format is not forced, then check the accept header
	if ctx.Accepts(fiber.MIMEApplicationJSON) == fiber.MIMEApplicationJSON {
		return ctx.JSON(fiber.Map{
			"message": message,
		})
	}

	if ctx.Accepts(fiber.MIMETextPlain) == fiber.MIMETextPlain {
		ctx.Response().Header.SetContentType(fiber.MIMETextPlain)
		return ctx.SendString(message)
	}

	// default response, html error page
	ctx.Response().Header.SetContentType(fiber.MIMETextHTML)
	return ctx.SendString(ErrorPage(fmt.Sprintf("ERROR %d", respCode), fmt.Sprintf("%d", respCode), message, "/"))
}

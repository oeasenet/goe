package webresult

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe"
	"go.uber.org/zap"
)

type WebResult struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func InvalidParam(msg ...string) *fiber.Error {
	if len(msg) > 0 && len(msg[0]) > 0 {
		return fiber.NewError(fiber.StatusBadRequest, msg[0])
	}
	return fiber.NewError(fiber.StatusBadRequest, "invalid request data")
}

func Unauthorized(msg ...string) *fiber.Error {
	if len(msg) > 0 && len(msg[0]) > 0 {
		return fiber.NewError(fiber.StatusUnauthorized, msg[0])
	}
	return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
}

func SendSucceed(ctx fiber.Ctx, data ...any) error {
	result := &WebResult{
		Message: "success",
	}
	if len(data) > 0 && data[0] != nil {
		result.Data = data[0]
	}
	return ctx.Status(fiber.StatusOK).JSON(result)
}

func SendFailed(ctx fiber.Ctx, msg string, data ...any) error {
	if msg == "" {
		msg = "operation failed"
	}
	result := &WebResult{
		Message: msg,
	}
	if len(data) > 0 && data[0] != nil {
		result.Data = data[0]
	}
	return ctx.Status(fiber.StatusBadRequest).JSON(result)
}

func NotFound(msg ...string) error {
	if len(msg) > 0 && len(msg[0]) > 0 {
		return fiber.NewError(fiber.StatusNotFound, msg[0])
	}
	return fiber.NewError(fiber.StatusNotFound, "resource not found")
}

func SystemBusy(err ...error) error {
	if len(err) > 0 && err[0] != nil {
		goe.UseLog().GetZapSugarLogger().WithOptions(zap.AddCallerSkip(0)).Error(err)
	}
	return fiber.NewError(fiber.StatusInternalServerError, "system busy")
}

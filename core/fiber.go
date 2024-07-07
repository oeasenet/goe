package core

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/validate"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe/contracts"
)

type GoeFiber struct {
	goeConfig *GoeConfig
	fiberApp  *fiber.App
	logger    contracts.Logger
}

// NewGoeFiber creates a new GoeFiber instance.
func NewGoeFiber(goeConfig *GoeConfig, l contracts.Logger) *GoeFiber {
	gf := &GoeFiber{
		goeConfig: goeConfig,
		logger:    l,
	}
	gf.initValidator()
	fiberApp := fiber.New(fiber.Config{
		ServerHeader:            goeConfig.Http.ServerHeader,
		StrictRouting:           false,
		CaseSensitive:           false,
		Immutable:               false,
		UnescapePath:            false,
		BodyLimit:               goeConfig.Http.BodyLimit,
		StreamRequestBody:       true,
		Concurrency:             goeConfig.Http.Concurrency,
		ProxyHeader:             goeConfig.Http.ProxyHeader,
		ErrorHandler:            gf.GoeFiberErrorHandler,
		AppName:                 gf.goeConfig.App.Name,
		ReduceMemoryUsage:       goeConfig.Http.ReduceMemory,
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
		XMLEncoder:              xml.Marshal,
		EnableTrustedProxyCheck: goeConfig.Http.TrustProxyCheck,
		TrustedProxies:          goeConfig.Http.TrustProxies,
		EnableIPValidation:      goeConfig.Http.IPValidation,
		ColorScheme:             fiber.DefaultColors,
		StructValidator:         gf.newFiberBindValidator(),
	})
	gf.fiberApp = fiberApp
	return gf
}

func (gf *GoeFiber) App() *fiber.App {
	return gf.fiberApp
}

func (gf *GoeFiber) CreateFiberApp(appName ...string) *fiber.App {
	if len(appName) == 0 || appName[0] == "" {
		appName[0] = gf.goeConfig.App.Name
	} else {
		gf.goeConfig.App.Name = appName[0]
	}
	return fiber.New(fiber.Config{
		ServerHeader:            gf.goeConfig.Http.ServerHeader,
		StrictRouting:           false,
		CaseSensitive:           false,
		Immutable:               false,
		UnescapePath:            false,
		BodyLimit:               gf.goeConfig.Http.BodyLimit,
		StreamRequestBody:       true,
		Concurrency:             gf.goeConfig.Http.Concurrency,
		ProxyHeader:             gf.goeConfig.Http.ProxyHeader,
		ErrorHandler:            gf.GoeFiberErrorHandler,
		AppName:                 appName[0],
		ReduceMemoryUsage:       gf.goeConfig.Http.ReduceMemory,
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
		XMLEncoder:              xml.Marshal,
		EnableTrustedProxyCheck: gf.goeConfig.Http.TrustProxyCheck,
		TrustedProxies:          gf.goeConfig.Http.TrustProxies,
		EnableIPValidation:      gf.goeConfig.Http.IPValidation,
		ColorScheme:             fiber.DefaultColors,
		StructValidator:         gf.newFiberBindValidator(),
	})
}

func (gf *GoeFiber) initValidator() {
	validate.Config(func(opt *validate.GlobalOption) {
		opt.ValidateTag = "v"
		opt.MessageTag = "m"
		opt.SkipOnEmpty = true
	})
	validate.AddValidator("id", func(val interface{}) bool {
		if val == nil {
			return false
		}
		if s, err := strutil.ToString(val); err != nil {
			return false
		} else {
			if _, err := primitive.ObjectIDFromHex(s); err != nil {
				return false
			}
		}
		return true
	})
	validate.AddGlobalMessages(map[string]string{
		"id": "{field} is not a valid ID",
	})
}

func (gf *GoeFiber) GoeFiberErrorHandler(ctx fiber.Ctx, err error) error {
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
	if ctx.Query("format") == "json" {
		ctx.Response().Header.SetContentType(fiber.MIMETextPlain)
		return ctx.SendString(message)
	}

	// If the format is not forced, then check the accept header
	if ctx.Accepts() == fiber.MIMEApplicationJSON {
		return ctx.JSON(fiber.Map{
			"message": message,
		})
	}

	if ctx.Accepts() == fiber.MIMETextPlain {
		ctx.Response().Header.SetContentType(fiber.MIMETextPlain)
		return ctx.SendString(message)
	}

	// default response, html error page
	ctx.Response().Header.SetContentType(fiber.MIMETextHTML)
	return ctx.SendString(ErrorPage(fmt.Sprintf("ERROR %d", respCode), fmt.Sprintf("%d", respCode), message, "/"))
}

type fiberBindValidator struct {
}

func (gf *GoeFiber) newFiberBindValidator() *fiberBindValidator {
	return &fiberBindValidator{}
}

func (f *fiberBindValidator) Validate(out any) error {
	v := validate.Struct(out)
	if !v.Validate() {
		return fiber.NewError(fiber.StatusBadRequest, v.Errors.One())
	}
	return nil
}

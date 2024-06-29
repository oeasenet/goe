package validation

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/validate"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
)

type FiberValidator struct {
	once sync.Once
}

func NewFiberValidator() *FiberValidator {
	fv := &FiberValidator{}
	fv.once.Do(func() {
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
	})
	return fv
}

func (f *FiberValidator) Handler(out any) error {
	v := validate.Struct(out)
	if !v.Validate() {
		return fiber.NewError(fiber.StatusBadRequest, v.Errors.One())
	}
	return nil
}

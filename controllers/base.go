package controllers

import (
	"github.com/sy264115809/logrush"
	validator "gopkg.in/go-playground/validator.v9"
)

// Base controller packages the common methods.
type Base struct {
	L *logrush.Logger

	validateFunc func(interface{}) error
}

var defaultValidator = validator.New().Struct

// New instances a new Base controller object.
func New() *Base {
	return &Base{
		L:            logrush.StandardLogger(),
		validateFunc: defaultValidator,
	}
}

// Copy replicates a new Base controller with given name.
func (c *Base) Copy(name string) *Base {
	return &Base{
		L:            c.L.Copy(name),
		validateFunc: c.validateFunc,
	}
}

// SetLogger sets logger of base controller.
func (c *Base) SetLogger(logger *logrush.Logger) *Base {
	c.L = logger
	return c
}

// SetValidateFunc allows to set a validate function for request binding validation.
// You can set it to nil to disable validation.
func (c *Base) SetValidateFunc(v func(interface{}) error) *Base {
	c.validateFunc = v
	return c
}

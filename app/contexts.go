package app

import "github.com/labstack/echo"
import "net/http"

func handleError(ctx echo.Context, r error) error {
	switch r := r.(type) {
	case *Error:
		r.ID = ctx.Response().Header().Get("X-Request-ID")
		return ctx.JSON(r.Status, r)
	default:
		return handleError(ctx, createInternalError(r))
	}
}

// UserShowContext provides the user show action context
type UserShowContext struct {
	context       echo.Context
	UserID        string
	CurrentUserID string
}

// NewUserShowContext parses the incoming request and create context
func NewUserShowContext(ctx echo.Context) (*UserShowContext, error) {
	var err error
	rctx := UserShowContext{context: ctx}
	rctx.UserID = ctx.Param("userID")
	return &rctx, err
}

// NotFound sends a HTTP response
func (ctx *UserShowContext) NotFound() error {
	return ctx.context.NoContent(http.StatusNotFound)
}

// InternalServerError sends a HTTP response
func (ctx *UserShowContext) InternalServerError(r error) error {
	return handleError(ctx.context, r)
}

// OK sends a HTTP response
func (ctx *UserShowContext) OK(r *UserView) error {
	return ctx.context.JSON(http.StatusOK, r)
}

// OKMe send a HTTP response
func (ctx *UserShowContext) OKMe(r *UserMeView) error {
	return ctx.context.JSON(http.StatusOK, r)
}

// UserUpdateContext provides the user update action context
type UserUpdateContext struct {
	context echo.Context
	UserID  string
	Payload *UserPayload
}

// NewUserUpdateContext parses the incoming request and create context
func NewUserUpdateContext(ctx echo.Context) (*UserUpdateContext, error) {
	var err error
	rctx := UserUpdateContext{context: ctx}
	rctx.UserID = ctx.Param("userID")
	return &rctx, err
}

// NoContent sends a HTTP response
func (ctx *UserUpdateContext) NoContent() error {
	return ctx.context.NoContent(http.StatusNoContent)
}

// BadRequest sends a HTTP response
func (ctx *UserUpdateContext) BadRequest(r error) error {
	return handleError(ctx.context, r)
}

// NotFound sends a HTTP response
func (ctx *UserUpdateContext) NotFound() error {
	return ctx.context.NoContent(http.StatusNotFound)
}

// InternalServerError sends a HTTP response
func (ctx *UserUpdateContext) InternalServerError(r error) error {
	return handleError(ctx.context, r)
}
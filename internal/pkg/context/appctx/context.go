package appctx

import (
	"context"
	"net/http"

	"github.com/acoshift/middleware"
	"github.com/moonrhythm/session"

	"github.com/acoshift/acourse/internal/pkg/app"
	"github.com/acoshift/acourse/internal/pkg/user"
)

type (
	userKey struct{}
	sessKey struct{}
)

// session id
const sessName = "sess"

// NewUserContext creates new context with user
func NewUserContext(ctx context.Context, user *user.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

// GetUser gets user from context
func GetUser(ctx context.Context) *user.User {
	x, _ := ctx.Value(userKey{}).(*user.User)
	return x
}

func newSessionContext(ctx context.Context, s *session.Session) context.Context {
	return context.WithValue(ctx, sessKey{}, s)
}

// getSession gets session from context
func getSession(ctx context.Context) *session.Session {
	return ctx.Value(sessKey{}).(*session.Session)
}

// Middleware is appctx middleware
func Middleware() middleware.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			sess, err := session.Get(ctx, sessName)
			if err != nil {
				panic(err)
			}
			ctx = newSessionContext(ctx, sess)

			userID := GetUserID(ctx)
			if userID != "" {
				u, err := user.Get(ctx, userID)
				if err == app.ErrNotFound {
					u = &user.User{
						ID:       userID,
						Username: userID,
					}
				}
				if err != nil {
					panic(err)
				}
				ctx = NewUserContext(ctx, u)
			}

			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	}
}

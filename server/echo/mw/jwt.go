package mw

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/sevenNt/ares/server/echo"
)

// DefaultJWT is the default JWT middleware.
var DefaultJWT = JWT{
	SigningMethod: "HS256",
	ContextKey:    "user",
	TokenLookup:   "header:" + echo.HeaderAuthorization,
	AuthScheme:    "bearer",
	Claims:        jwt.MapClaims{},
}

// JWT defines the config for JWT middleware.
type JWT struct {
	Base
	SigningKey    interface{} // Signing key to validate token, required
	SigningMethod string
	ContextKey    string
	Claims        jwt.Claims
	TokenLookup   string
	AuthScheme    string
	KeyFunc       jwt.Keyfunc
}

type jwtExtractor func(*echo.Context) (string, error)

// Func implements Middleware interface.
func (j JWT) Func() echo.MiddlewareFunc {
	if j.Claims == nil {
		j.Claims = DefaultJWT.Claims
	}
	if j.TokenLookup == "" {
		j.TokenLookup = DefaultJWT.TokenLookup
	}
	if j.AuthScheme == "" {
		j.AuthScheme = DefaultJWT.AuthScheme
	}
	if j.SigningMethod == "" {
		j.SigningMethod = DefaultJWT.SigningMethod
	}
	if j.SigningKey == nil {
		j.SigningKey = DefaultJWT.SigningKey
	}
	j.KeyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != j.SigningMethod {
			return nil, fmt.Errorf("Unexpected jwt signing method=%v", t.Header["alg"])
		}
		return j.SigningKey, nil
	}

	// Initialize
	parts := strings.Split(j.TokenLookup, ":")
	extractor := jwtFromHeader(parts[1], j.AuthScheme)
	switch parts[0] {
	case "query":
		extractor = jwtFromQuery(parts[1])
	case "cookie":
		extractor = jwtFromCookie(parts[1])
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			auth, err := extractor(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			token := new(jwt.Token)
			// Issue #647, #656
			if _, ok := j.Claims.(jwt.MapClaims); ok {
				token, err = jwt.Parse(auth, j.KeyFunc)
			} else {
				claims := reflect.ValueOf(j.Claims).Interface().(jwt.Claims)
				token, err = jwt.ParseWithClaims(auth, claims, j.KeyFunc)
			}
			if err == nil && token.Valid {
				// Store user information from token into context.
				c.Set(j.ContextKey, token)
				return next(c)
			}
			return echo.ErrUnauthorized
		}
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtExtractor {
	return func(c *echo.Context) (string, error) {
		auth := c.Request().Header().Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", errors.New("Missing or invalid jwt in the request header")
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the query string.
func jwtFromQuery(param string) jwtExtractor {
	return func(c *echo.Context) (string, error) {
		token := c.Query(param)
		if token == "" {
			return "", errors.New("Missing jwt in the query string")
		}
		return token, nil
	}
}

// jwtFromCookie returns a `jwtExtractor` that extracts token from the named cookie.
func jwtFromCookie(name string) jwtExtractor {
	return func(c *echo.Context) (string, error) {
		cookie, err := c.Cookie(name)
		if err != nil {
			return "", errors.New("Missing jwt in the cookie")
		}
		return cookie.Value, nil
	}
}

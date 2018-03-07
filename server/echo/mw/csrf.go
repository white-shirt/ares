package mw

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sevenNt/ares/server/echo"
	"github.com/sevenNt/ares/util"
)

// DefaultCSRF is the default CSRF middleware.
var DefaultCSRF = CSRF{
	TokenLength:  32,
	TokenLookup:  "header:" + echo.HeaderXCSRFToken,
	ContextKey:   "csrf",
	CookieName:   "_csrf",
	CookieMaxAge: 86400,
}

type csrfTokenExtractor func(*echo.Context) (string, error)

// CSRF csrf
type CSRF struct {
	Base
	// TokenLength is the length of the generated token.
	TokenLength uint8 `json:"token_length"`
	// Optional. Default value 32.

	// TokenLookup is a string in the form of "<source>:<key>" that is used
	// to extract token from the request.
	// Optional. Default value "header:X-CSRF-Token".
	// Possible values:
	// - "header:<name>"
	// - "form:<name>"
	// - "query:<name>"
	TokenLookup string `json:"token_lookup"`

	// Context key to store generated CSRF token into context.
	// Optional. Default value "csrf".
	ContextKey string `json:"context_key"`

	// Name of the CSRF cookie. This cookie will store CSRF token.
	// Optional. Default value "csrf".
	CookieName string `json:"cookie_name"`

	// Domain of the CSRF cookie.
	// Optional. Default value none.
	CookieDomain string `json:"cookie_domain"`

	// Path of the CSRF cookie.
	// Optional. Default value none.
	CookiePath string `json:"cookie_path"`

	// Max age (in seconds) of the CSRF cookie.
	// Optional. Default value 86400 (24hr).
	CookieMaxAge int `json:"cookie_max_age"`

	// Indicates if CSRF cookie is secure.
	// Optional. Default value false.
	CookieSecure bool `json:"cookie_secure"`

	// Indicates if CSRF cookie is HTTP only.
	// Optional. Default value false.
	CookieHTTPOnly bool `json:"cookie_http_only"`
}

// Func implements Middleware interface
func (m CSRF) Func() echo.MiddlewareFunc {
	if m.TokenLength == 0 {
		m.TokenLength = DefaultCSRF.TokenLength

	}
	if m.TokenLookup == "" {
		m.TokenLookup = DefaultCSRF.TokenLookup

	}
	if m.ContextKey == "" {
		m.ContextKey = DefaultCSRF.ContextKey

	}
	if m.CookieName == "" {
		m.CookieName = DefaultCSRF.CookieName

	}
	if m.CookieMaxAge == 0 {
		m.CookieMaxAge = DefaultCSRF.CookieMaxAge

	}

	// Initialize
	parts := strings.Split(m.TokenLookup, ":")
	extractor := csrfTokenFromHeader(parts[1])
	switch parts[0] {
	case "form":
		extractor = csrfTokenFromForm(parts[1])
	case "query":
		extractor = csrfTokenFromQuery(parts[1])
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			k, err := c.Cookie(m.CookieName)
			token := ""

			if err != nil {
				// Generate token
				token = util.RandString(m.TokenLength)
			} else {
				// Reuse token
				token = k.Value
			}

			switch req.Method {
			case echo.GET, echo.HEAD, echo.OPTIONS, echo.TRACE:
			default:
				// Validate token only for requests which are not defined as 'safe' by RFC7231
				clientToken, err := extractor(c)
				if err != nil {
					return err
				}
				if !validateCSRFToken(token, clientToken) {
					return echo.NewHTTPError(http.StatusForbidden, "csrf token is invalid")
				}
			}

			// Set CSRF cookie
			cookie := new(http.Cookie)
			cookie.Name = m.CookieName
			cookie.Value = token
			if m.CookiePath != "" {
				cookie.Path = m.CookiePath
			}
			if m.CookieDomain != "" {
				cookie.Domain = m.CookieDomain
			}
			cookie.Expires = time.Now().Add(time.Duration(m.CookieMaxAge) * time.Second)
			cookie.Secure = m.CookieSecure
			cookie.HttpOnly = m.CookieHTTPOnly
			c.SetCookie(cookie)

			// Store token in the context
			c.Set(m.ContextKey, token)

			// Protect clients from caching the response
			c.Response().Header().Add(echo.HeaderVary, echo.HeaderCookie)
			return next(c)
		}
	}
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided request header.
func csrfTokenFromHeader(header string) csrfTokenExtractor {
	return func(c *echo.Context) (string, error) {
		return c.Request().Header().Get(header), nil
	}
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided form parameter.
func csrfTokenFromForm(param string) csrfTokenExtractor {
	return func(c *echo.Context) (string, error) {
		token := c.FormValue(param)
		if token == "" {
			return "", errors.New("empty csrf token in form param")
		}
		return token, nil
	}
}

// csrfTokenFromQuery returns a `csrfTokenExtractor` that extracts token from the
// provided query parameter.
func csrfTokenFromQuery(param string) csrfTokenExtractor {
	return func(c *echo.Context) (string, error) {
		token := c.QueryString(param, "")
		if token == "" {
			return "", errors.New("empty csrf token in query param")
		}
		return token, nil
	}
}

func validateCSRFToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}

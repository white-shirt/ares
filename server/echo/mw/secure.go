package mw

import (
	"fmt"

	"github.com/sevenNt/ares/server/echo"
)

type ( // Secure defines the config for Recovery middleware.
	Secure struct {
		Base

		// XSSProtection provides protection against cross-site scripting attack (XSS)
		// by setting the `X-XSS-Protection` header.
		// Optional. Default value "1; mode=block".
		XSSProtection string `json:"xss_protection"`

		// ContentTypeNosniff provides protection against overriding Content-Type
		// header by setting the `X-Content-Type-Options` header.
		// Optional. Default value "nosniff".
		ContentTypeNosniff string `json:"content_type_nosniff"`

		// XFrameOptions can be used to indicate whether or not a browser should
		// be allowed to render a page in a <frame>, <iframe> or <object> .
		// Sites can use this to avoid clickjacking attacks, by ensuring that their
		// content is not embedded into other sites.provides protection against
		// clickjacking.
		// Optional. Default value "SAMEORIGIN".
		// Possible values:
		// - "SAMEORIGIN" - The page can only be displayed in a frame on the same origin as the page itself.
		// - "DENY" - The page cannot be displayed in a frame, regardless of the site attempting to do so.
		// - "ALLOW-FROM uri" - The page can only be displayed in a frame on the specified origin.
		XFrameOptions string `json:"x_frame_options"`

		// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how
		// long (in seconds) browsers should remember that this site is only to
		// be accessed using HTTPS. This reduces your exposure to some SSL-stripping
		// man-in-the-middle (MITM) attacks.
		// Optional. Default value 0.
		HSTSMaxAge int `json:"hsts_max_age"`

		// HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`
		// header, excluding all subdomains from security policy. It has no effect
		// unless HSTSMaxAge is set to a non-zero value.
		// Optional. Default value false.
		HSTSExcludeSubdomains bool `json:"hsts_exclude_subdomains"`

		// ContentSecurityPolicy sets the `Content-Security-Policy` header providing
		// security against cross-site scripting (XSS), clickjacking and other code
		// injection attacks resulting from execution of malicious content in the
		// trusted web page context.
		// Optional. Default value "".
		ContentSecurityPolicy string `json:"content_security_policy"`
	}
)

var (
	// DefaultSecure is the default instance of Secure.
	DefaultSecure = Secure{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "SAMEORIGIN",
	}
)

// Func implements Middleware interface.
func (s Secure) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			res := c.Response()

			if s.XSSProtection != "" {
				res.Header().Set(echo.HeaderXXSSProtection, s.XSSProtection)
			}
			if s.ContentTypeNosniff != "" {
				res.Header().Set(echo.HeaderXContentTypeOptions, s.ContentTypeNosniff)
			}
			if s.XFrameOptions != "" {
				res.Header().Set(echo.HeaderXFrameOptions, s.XFrameOptions)
			}
			if (req.IsTLS() || (req.Header().Get(echo.HeaderXForwardedProto) == "https")) && s.HSTSMaxAge != 0 {
				subdomains := ""
				if !s.HSTSExcludeSubdomains {
					subdomains = "; includeSubdomains"
				}
				res.Header().Set(echo.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", s.HSTSMaxAge, subdomains))
			}
			if s.ContentSecurityPolicy != "" {
				res.Header().Set(echo.HeaderContentSecurityPolicy, s.ContentSecurityPolicy)
			}
			return next(c)
		}
	}
}

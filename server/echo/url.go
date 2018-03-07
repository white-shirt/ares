package echo

import "net/url"

// URL scheme://[userinfo@]host/path[?query][#fragment]
type URL struct {
	*url.URL
	query url.Values
}

func newURL(url *url.URL) *URL {
	return &URL{
		URL:   url,
		query: nil,
	}
}

// Path returns escaped path.
func (u *URL) Path() string {
	return u.URL.EscapedPath()
}

// Query return query by name.
func (u *URL) Query(name string) string {
	if u.query == nil {
		u.query = u.URL.Query()
	}

	return u.query.Get(name)
}

// QueryS return queries by name.
// TODO framework-specific
// "?list_a=1&list_a=2&list_a=3&list_b[]=1&list_b[]=2&list_b[]=3&list_c=1,2,3"
// would be parsed as:
// list_a = [1,2,3]
// list_b, not support
// list_c = "1,2,3"
func (u *URL) QueryS(name string) []string {
	if u.query == nil {
		u.query = u.URL.Query()
	}

	return u.query[name]
}

// QueryAll returns all queries.
func (u *URL) QueryAll() url.Values {
	if u.query == nil {
		u.query = u.URL.Query()
	}

	return u.query
}

func (u *URL) reset(url *url.URL) {
	u.URL = url
	u.query = nil
}

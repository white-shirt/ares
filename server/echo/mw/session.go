package mw

import (
	"log"
	"net/http"

	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
	"github.com/sevenNt/ares/server/echo"
)

const (
	// DefaultKey is session default key in context.
	DefaultKey  = "session_key"
	errorFormat = "[sessions] ERROR! %s\n"
)

// var store, _ = NewRedisStore(10, "tcp", "10.1.61.126:6386", "VuE9ZHL2HoC7952ZKpkcx23UqWc4Dr", []byte("5c0e8d1f081970ad"))
// var store = NewFileSystemStoreStore(".session", []byte("5c0e8d1f081970ad"))
var store = NewCookieStore([]byte("5c0e8d1f081970ad"))

// DefaultSession is an instance of Session.
var DefaultSession = &Session{
	name:  "mysession",
	store: store,
}

// Store stores sessions.
type Store interface {
	sessions.Store
	Options(SessionOptions)
	MaxAge(int)
}

// SessionConfig offers a declarative way to construct a session.
type SessionConfig struct {
	Type   string
	Name   string
	Path   string
	Size   int
	Addr   string
	Pwd    string
	Secret string
}

// NewSession constructs a session.
func NewSession(cfg SessionConfig) *Session {
	var err error
	var store RedisStore
	switch cfg.Type {
	case "redis":
		store, err = NewRedisStore(cfg.Size, "tcp", cfg.Addr, cfg.Pwd, []byte(cfg.Secret))
		break
	case "file":
		store = NewFileSystemStoreStore(cfg.Path, []byte(cfg.Secret))
		break
	case "cookie":
		store = NewCookieStore([]byte(cfg.Secret))
		break
	default:
		panic("no session type")
	}
	if err != nil {
		panic(err)
	}
	return &Session{
		name:  cfg.Name,
		store: store,
	}
}

// SessionOptions stores configuration for a Session or Session store.
// Fields are a subset of http.Cookie fields.
type SessionOptions struct {
	Path   string
	Domain string
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int
	Secure   bool
	HTTPOnly bool
}

// Func implements Middleware interface
func (s Session) Func() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set(DefaultKey, &Session{
				name:    s.name,
				request: c.Request().Request,
				store:   s.store,
				session: nil,
				written: false,
				writer:  c.Response().ResponseWriter,
			})
			return next(c)
		}
	}
}

// Session stores the values and optional configuration for a Session.
type Session struct {
	Base
	name    string
	request *http.Request
	store   Store
	session *sessions.Session
	written bool
	writer  http.ResponseWriter
}

// Get gets value by key.
func (s *Session) Get(key interface{}) interface{} {
	return s.Session().Values[key]
}

// Set sets value with key.
func (s *Session) Set(key interface{}, val interface{}) {
	s.Session().Values[key] = val
	s.written = true
}

// Delete deletes value by key.
func (s *Session) Delete(key interface{}) {
	delete(s.Session().Values, key)
	s.written = true
}

// Clear clears values.
func (s *Session) Clear() {
	for key := range s.Session().Values {
		s.Delete(key)
	}
}

// AddFlash adds a flash message to the session.
func (s *Session) AddFlash(value interface{}, vars ...string) {
	s.Session().AddFlash(value, vars...)
	s.written = true
}

// Flashes returns a slice of flash messages from the session.
func (s *Session) Flashes(vars ...string) []interface{} {
	s.written = true
	return s.Session().Flashes(vars...)
}

// Options sets session options.
func (s *Session) Options(options SessionOptions) {
	s.Session().Options = &sessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
	}
}

// Save is a convenience method to save this session.
func (s *Session) Save() error {
	if s.Written() {
		e := s.Session().Save(s.request, s.writer)
		if e == nil {
			s.written = false
		}
		return e
	}
	return nil
}

// Session returns session from store.
func (s *Session) Session() *sessions.Session {
	if s.session == nil {
		var err error
		s.session, err = s.store.Get(s.request, s.name)
		if err != nil {
			log.Printf(errorFormat, err)
		}
	}
	return s.session
}

// Written returns session written status.
func (s *Session) Written() bool {
	return s.written
}

// SessionWithCtx offers shortcut to Session
func SessionWithCtx(c *echo.Context) *Session {
	session := c.Get(DefaultKey)
	if session == nil {
		return nil
	}
	return c.Get(DefaultKey).(*Session)
}

// CookieStore offers cookie store of session.
type CookieStore interface {
	Store
}

// NewCookieStore returns cookie store.
// Keys are defined in pairs to allow key rotation, but the common case is to set a single
// authentication key and optionally an encryption key.
//
// The first key in a pair is used for authentication and the second for encryption. The
// encryption key can be set to nil or omitted in the last pair, but the authentication key
// is required in all pairs.
//
// It is recommended to use an authentication key with 32 or 64 bytes. The encryption key,
// if set, must be either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256 modes.
func NewCookieStore(keyPairs ...[]byte) CookieStore {
	return &cookieStore{sessions.NewCookieStore(keyPairs...)}
}

type cookieStore struct {
	*sessions.CookieStore
}

func (c *cookieStore) Options(options SessionOptions) {
	c.CookieStore.Options = &sessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
	}
}

func (c *cookieStore) MaxAge(age int) {
	c.CookieStore.MaxAge(age)
}

// FileSystemStore offers file store of session.
type FileSystemStore interface {
	Store
	MaxLength(int)
}

// NewFileSystemStoreStore returns file store.
// The path argument is the directory where sessions will be saved. If empty
// it will use os.TempDir().
//
// Keys are defined in pairs to allow key rotation, but the common case is to set a single
// authentication key and optionally an encryption key.
//
// The first key in a pair is used for authentication and the second for encryption. The
// encryption key can be set to nil or omitted in the last pair, but the authentication key
// is required in all pairs.
//
// It is recommended to use an authentication key with 32 or 64 bytes. The encryption key,
// if set, must be either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256 modes.
func NewFileSystemStoreStore(path string, keyPairs ...[]byte) FileSystemStore {
	return &filesystemStore{sessions.NewFilesystemStore(path, keyPairs...)}
}

type filesystemStore struct {
	*sessions.FilesystemStore
}

func (s *filesystemStore) Options(options SessionOptions) {
	s.FilesystemStore.Options = &sessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
	}
}

// MaxAge sets the maximum age for the store and the underlying cookie
// implementation. Individual sessions can be deleted by setting SessionOptions.MaxAge
// = -1 for that session.
func (s *filesystemStore) MaxAge(age int) {
	s.FilesystemStore.MaxAge(age)
}

// MaxLength restricts the maximum length of new sessions to l.
// If l is 0 there is no limit to the size of a session, use with caution.
// The default for a new FilesystemStore is 4096.
func (s *filesystemStore) MaxLength(l int) {
	s.FilesystemStore.MaxLength(l)
}

// RedisStore offers redis store of session.
type RedisStore interface {
	Store
}

// NewRedisStore returns redis store.
// size: maximum number of idle connections.
// network: tcp or udp
// address: host:port
// password: redis-password
// Keys are defined in pairs to allow key rotation, but the common case is to set a single
// authentication key and optionally an encryption key.
//
// The first key in a pair is used for authentication and the second for encryption. The
// encryption key can be set to nil or omitted in the last pair, but the authentication key
// is required in all pairs.
//
// It is recommended to use an authentication key with 32 or 64 bytes. The encryption key,
// if set, must be either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256 modes.
func NewRedisStore(size int, network, address, password string, keyPairs ...[]byte) (RedisStore, error) {
	store, err := redistore.NewRediStore(size, network, address, password, keyPairs...)
	if err != nil {
		return nil, err
	}
	return &redisStore{store}, nil
}

type redisStore struct {
	*redistore.RediStore
}

func (c *redisStore) Options(options SessionOptions) {
	c.RediStore.Options = &sessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
	}
}

// MaxAge restricts the maximum age, in seconds, of the session record
// both in database and a browser. This is to change session storage configuration.
// If you want just to remove session use your session `s` object and change it's
// `SessionOptions.MaxAge` to -1, as specified in
//    http://godoc.org/github.com/gorilla/sessions#Options
//
// Default is the one provided by github.com/boj/redistore package value - `sessionExpire`.
// Set it to 0 for no restriction.
// Because we use `MaxAge` also in SecureCookie crypting algorithm you should
// use this function to change `MaxAge` value.
func (c *redisStore) MaxAge(age int) {
	c.RediStore.SetMaxAge(age)
}

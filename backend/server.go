package backend

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"euphoria.io/heim/proto"
	"euphoria.io/heim/proto/security"
	"euphoria.io/heim/templates"
	"euphoria.io/scope"

	gorillactx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
)

const cookieKeySize = 32

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"heim1"},
	CheckOrigin:     checkOrigin,
}

type Server struct {
	ID            string
	Era           string
	r             *mux.Router
	heim          *proto.Heim
	b             proto.Backend
	kms           security.KMS
	pageModTime   time.Time
	pageTemplater *templates.Templater
	staticPath    string
	sc            *securecookie.SecureCookie
	rootCtx       scope.Context

	allowRoomCreation     bool
	allowAPI              bool
	showAllRooms          bool
	newAccountMinAgentAge time.Duration
	roomEntryMinAgentAge  time.Duration
	setInsecureCookies    bool

	m sync.Mutex

	agentIDGenerator func() ([]byte, error)
}

func NewServer(heim *proto.Heim, id, era string) (*Server, error) {
	mime.AddExtensionType(".map", "application/json")

	cookieSecret, err := heim.Cluster.GetSecret(heim.KMS, "cookie", cookieKeySize)
	if err != nil {
		return nil, fmt.Errorf("error coordinating shared cookie secret: %s", err)
	}

	s := &Server{
		ID:            id,
		Era:           era,
		heim:          heim,
		b:             heim.Backend,
		kms:           heim.KMS,
		pageModTime:   time.Now(),
		pageTemplater: heim.PageTemplater,
		staticPath:    heim.StaticPath,
		sc:            securecookie.New(cookieSecret, nil),
		rootCtx:       heim.Context,
		allowAPI:      true,
	}
	s.route()
	return s, nil
}

func (s *Server) AllowRoomCreation(allow bool)            { s.allowRoomCreation = allow }
func (s *Server) AllowAPI(allow bool)                     { s.allowAPI = allow }
func (s *Server) ShowAllRooms(enable bool)                { s.showAllRooms = enable }
func (s *Server) NewAccountMinAgentAge(age time.Duration) { s.newAccountMinAgentAge = age }
func (s *Server) RoomEntryMinAgentAge(age time.Duration)  { s.roomEntryMinAgentAge = age }
func (s *Server) SetInsecureCookies(allow bool)           { s.setInsecureCookies = allow }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

type gzipResponseWriter struct {
	http.ResponseWriter
	cache bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	header := w.Header()
	header.Set("Content-Encoding", "gzip")
	header.Add("Vary", "Accept-Encoding")
	if w.cache {
		header.Add("Cache-Control", "max-age=604800")
	}
	w.ResponseWriter.WriteHeader(code)
}

func (s *Server) serveGzippedFile(w http.ResponseWriter, r *http.Request, filename string, cache bool) {
	dir := http.Dir(s.staticPath)
	var err error
	var f http.File
	gzipped := false

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		f, err = dir.Open(filename + ".gz")
		if err != nil {
			f = nil
		} else {
			gzipped = true
		}
	}

	if f == nil {
		f, err = dir.Open(filename)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	name := d.Name()
	if gzipped {
		name = strings.TrimSuffix(name, ".gz")
		w = &gzipResponseWriter{
			ResponseWriter: w,
			cache:          cache,
		}
	}

	http.ServeContent(w, r, name, d.ModTime(), f)
}

type hijackResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
}

func instrumentSocketHandlerFunc(name string, handler http.HandlerFunc) http.HandlerFunc {
	type hijackerKey int
	var k hijackerKey

	loadHijacker := func(w http.ResponseWriter, r *http.Request) {
		if hj, ok := gorillactx.GetOk(r, k); ok {
			w = &hijackResponseWriter{ResponseWriter: w, Hijacker: hj.(http.Hijacker)}
		}
		handler(w, r)
	}

	promHandler := prometheus.InstrumentHandlerFunc(name, loadHijacker)

	saveHijacker := func(w http.ResponseWriter, r *http.Request) {
		if hj, ok := w.(http.Hijacker); ok {
			gorillactx.Set(r, k, hj)
		}
		promHandler(w, r)
	}

	return saveHijacker
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]

	// If no Origin header was given, accept.
	if len(origin) == 0 {
		return true
	}

	// If Origin is "null", accept.
	if origin[0] == "null" {
		return true
	}

	// Try to parse Origin, and reject if there's an error.
	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}

	// If Origin matches any of these prefix/requested-host combinations, accept.
	for _, prefix := range []string{"", "www.", "alpha."} {
		if u.Host == prefix+r.Host {
			return true
		}
	}

	if u.Host == "localhost:8080" || u.Host == "euphoria.local:8080" {
		return true
	}

	// If we reach this point, reject the Origin.
	return false
}

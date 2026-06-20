package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"euphoria.leet.nu/lib/scope"

	"euphoria.leet.nu/heim/proto/logging"
	"euphoria.leet.nu/heim/templates"
)

const (
	RoomPage          = "room.html"
	ResetPasswordPage = "reset-password.html"
	VerifyEmailPage   = "verify-email.html"
	LibPage           = "libpage.html"
)

var PageScenarios = map[string]map[string]templates.TemplateTest{
	RoomPage: map[string]templates.TemplateTest{
		"default": templates.TemplateTest{
			Data: map[string]interface{}{"RoomName": "test"},
		},
	},
	ResetPasswordPage: map[string]templates.TemplateTest{
		"default": templates.TemplateTest{
			Data: map[string]interface{}{
				"Data": map[string]interface{}{
					"email":        "test@test.invalid",
					"confirmation": "confirmationcode",
				},
			},
		},
	},
	VerifyEmailPage: map[string]templates.TemplateTest{
		"default": templates.TemplateTest{
			Data: map[string]interface{}{
				"Data": map[string]interface{}{
					"email":        "test@test.invalid",
					"confirmation": "confirmationcode",
				},
			},
		},
	},
	LibPage: map[string]templates.TemplateTest{
		"default": templates.TemplateTest{
			Data: map[string]interface{}{
				"Name":        "magic",
				"GoImport":    "magic git https://example.com/magic",
				"ModulePath":  "magic",
				"VCS":         "sccs",
				"RepoURL":     "https://example.com/magic",
				"Description": "Magic solves all your problems on its own!",
			},
		},
		"subdir": templates.TemplateTest{
			Data: map[string]interface{}{
				"Name":        "magic/more",
				"GoImport":    "magic git https://example.com/magic more",
				"ModulePath":  "magic/more",
				"VCS":         "sccs",
				"RepoURL":     "https://example.com/magic",
				"SubdirName":  "more",
				"SubdirURL":   "https://example.com/magic/more",
				"Description": "Magic did not solve your problem? Try more magic!",
			},
		},
	},
}

func ValidatePageTemplates(templater templates.Templater) []error {
	return templates.ValidateTemplates(templater, PageScenarios)
}

func LoadPageTemplates(ctx scope.Context, path string) (templates.Templater, error) {
	pageTemplater := &templates.StandardTemplater{}
	if errs := pageTemplater.Load(path); errs != nil {
		return nil, errs[0]
	}
	if errs := ValidatePageTemplates(pageTemplater); errs != nil {
		for _, err := range errs {
			logging.Logger(ctx).Printf("error: %s\n", err)
		}
		return nil, fmt.Errorf("template validation failed: %s...", errs[0].Error())
	}
	return pageTemplater, nil
}

func (s *Server) servePage(name string, params map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	content, err := s.pageTemplater.Evaluate(name, params)
	if err != nil {
		switch err {
		case templates.ErrTemplateNotFound:
			s.serveErrorPage("page not found", http.StatusNotFound, w, r)
		default:
			s.serveErrorPage(err.Error(), http.StatusInternalServerError, w, r)
		}
		return
	}
	// TODO: figure out real modtime
	http.ServeContent(w, r, name, s.pageModTime, bytes.NewReader(content))
}

func (s *Server) serveJSONPage(name string, context map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	params, err := json.Marshal(context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.servePage(name, map[string]interface{}{"Data": string(params)}, w, r)
}

func (s *Server) serveErrorPage(message string, code int, w http.ResponseWriter, r *http.Request) {
	params := map[string]interface{}{"Message": message, "Code": code}
	content, err := s.pageTemplater.Evaluate("error.html", params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// http.ServeContent() would try to set its own status code, which generates an annoying warning,
	// and also try to process byte range requests, which cannot work well for error pages
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(code)
	if r.Method != "HEAD" {
		w.Write(content)
	}
}

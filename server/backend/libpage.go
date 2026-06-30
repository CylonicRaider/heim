package backend

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"euphoria.leet.nu/lib/scope"
	"gopkg.in/yaml.v2"

	"euphoria.leet.nu/heim/proto/logging"
)

type libPageSet struct {
	Providers []*libPageProvider      `yaml:"providers,omitempty"`
	Pages     map[string]*libPageDesc `yaml:"libs,omitempty"`
}

func LoadLibPages(ctx scope.Context, filename string) (*libPageSet, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	result := &libPageSet{}
	if err = yaml.Unmarshal(data, result); err != nil {
		return nil, err
	}
	errs := result.Init()
	if errs != nil {
		for _, err = range errs {
			logging.Logger(ctx).Printf("library page error: %s", err)
		}
		return nil, fmt.Errorf("loading library pages: %s...", errs[0])
	}
	return result, nil
}

func (lp *libPageSet) Init() []error {
	var errors []error
	for name, page := range lp.Pages {
		var importInfo *goImport
		for _, provider := range lp.Providers {
			if importInfo = provider.ParseURL(page.ModulePath, page.URL); importInfo != nil {
				break
			}
		}
		if importInfo == nil {
			errors = append(errors, fmt.Errorf("page %s URL %s does not match any provider",
				name, page.URL))
			continue
		}
		page.init(name, importInfo)
	}
	return errors
}

func (lp *libPageSet) Lookup(name string) map[string]interface{} {
	if lp == nil {
		return nil
	} else if page, ok := lp.Pages[name]; ok {
		return page.TemplateData
	} else {
		return nil
	}
}

type libPageProvider struct {
	// Pattern recognizes URLs as belonging to this provider and may have the following capturing groups:
	// - "repo": The repository URL for the go-import meta tag. If absent, the entire URL is the repository URL.
	// - "subdir": The name of the subdirectory inside the repository. If absent, there is no subdirectory.
	Regex regexp.Regexp `yaml:"pattern"`
	// VCS is the VCS name for Go imports
	VCS string `yaml:"vcs"`
}

func (p *libPageProvider) ParseURL(modulePath, url string) *goImport {
	match := p.Regex.FindStringSubmatch(url)
	if match == nil {
		return nil
	}
	repoIdx, subdirIdx := p.Regex.SubexpIndex("repo"), p.Regex.SubexpIndex("subdir")
	repoURL, subdir := url, ""
	if repoIdx != -1 {
		repoURL = match[repoIdx]
	}
	if subdirIdx != -1 {
		subdir = match[subdirIdx]
	}
	return &goImport{
		ModulePath: modulePath,
		VCS:        p.VCS,
		RepoURL:    repoURL,
		Subdir:     subdir,
	}
}

type goImport struct {
	ModulePath string
	VCS        string
	RepoURL    string
	Subdir     string
}

func (gi *goImport) String() string {
	parts := []string{gi.ModulePath, gi.VCS, gi.RepoURL}
	if gi.Subdir != "" {
		parts = append(parts, gi.Subdir)
	}
	return strings.Join(parts, " ")
}

type libPageDesc struct {
	ModulePath  string `yaml:"module"`
	URL         string `yaml:"url"`
	Description string `yaml:"desc"`

	// Filled in by init().
	Name         string
	TemplateData map[string]interface{}
}

func (d *libPageDesc) init(name string, importInfo *goImport) {
	d.Name = name
	d.TemplateData = map[string]interface{}{
		"Name":        name,
		"GoImport":    importInfo.String(),
		"ModulePath":  importInfo.ModulePath,
		"VCS":         importInfo.VCS,
		"RepoURL":     importInfo.RepoURL,
		"Description": d.Description,
	}
	if importInfo.Subdir != "" {
		d.TemplateData["SubdirName"] = importInfo.Subdir
		d.TemplateData["SubdirURL"] = d.URL
	}
}

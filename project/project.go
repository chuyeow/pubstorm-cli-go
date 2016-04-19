package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/nitrous-io/rise-cli-go/config"
)

type Project struct {
	Name string `json:"name"`
	Path string `json:"path"`

	DefaultDomainEnabled bool `json:"default_domain_enabled"`

	// TODO These 2 flags should be read from the API server (like
	// DefaultDomainEnabled).
	EnableStats bool `json:"enable_stats"`
	ForceHTTPS  bool `json:"force_https"`
}

var (
	projectNameRe = regexp.MustCompile(`(?m)(^[a-z0-9][a-z0-9\-]{1,61}[a-z0-9]$)`)

	ErrNameInvalidLength = errors.New("Name must have minimum 3 and maximum 63 characters")
	ErrNameInvalid       = errors.New("Name may only contain lowercase letters, numbers and hyphens, but may not begin or end with hyphens")

	ErrPathNotRelative = errors.New("Path must be relative to current working directory")
	ErrPathNotExist    = errors.New("Path does not exist")
	ErrPathNotDir      = errors.New("Path must be a directory")
)

func (p *Project) DefaultDomain() string {
	return fmt.Sprintf("%s.%s", p.Name, config.DefaultDomain)
}

// Validates name
func (p *Project) ValidateName() error {
	if len(p.Name) < 3 || len(p.Name) > 63 {
		return ErrNameInvalidLength
	}

	if !projectNameRe.MatchString(p.Name) {
		return ErrNameInvalid
	}

	return nil
}

// Validates path
func (p *Project) ValidatePath() error {
	if filepath.IsAbs(p.Path) {
		return ErrPathNotRelative
	}

	s, err := os.Stat(p.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrPathNotExist
		}
		return err
	}

	if !s.IsDir() {
		return ErrPathNotDir
	}

	return nil
}

// Save project settings to project json file
func (p *Project) Save() error {
	f, err := os.OpenFile(config.ProjectJSON, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(p)
}

func Load() (*Project, error) {
	f, err := os.Open(config.ProjectJSON)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	proj := Project{}
	if err = json.NewDecoder(f).Decode(&proj); err != nil {
		return nil, err
	}

	return &proj, nil
}

// Delete project json file
func (p *Project) Delete() error {
	return os.Remove(config.ProjectJSON)
}

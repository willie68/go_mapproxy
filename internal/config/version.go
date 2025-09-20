package config

import (
	"encoding/json"
	"fmt"

	"github.com/samber/do/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Version the version information
type Version struct {
	version string `json:"version"`
	commit  string `json:"commit"`
	date    string `json:"date"`
}

func Init(inj do.Injector) {
	ver := NewVersion()
	do.ProvideValue(inj, *ver)
}

// NewVersion creating a new version
func NewVersion() *Version {
	return &Version{
		version: version,
		commit:  commit,
		date:    date,
	}
}

// WithVersion setting the version information fluid
func (v *Version) WithVersion(version string) *Version {
	v.version = version
	return v
}

// WithCommit setting the commit information fluid
func (v *Version) WithCommit(commit string) *Version {
	v.commit = commit
	return v
}

// WithDate setting the date information fluid
func (v *Version) WithDate(date string) *Version {
	v.date = date
	return v
}

// Version return in the version information
func (v *Version) Version() string {
	return v.version
}

// Commit return in the commit information
func (v *Version) Commit() string {
	return v.commit
}

// Date return in the date information
func (v *Version) Date() string {
	return v.date
}

// JSON getting the version information as json
func (v *Version) JSON() (string, error) {
	ver := struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Date    string `json:"date"`
	}{
		Version: v.version,
		Commit:  v.commit,
		Date:    v.date,
	}
	js, err := json.Marshal(ver)
	return string(js), err
}

func (v *Version) String() string {
	return fmt.Sprintf("gomapproxy Version %s, commit: %s, builddate: %s", v.version, v.commit, v.date)
}

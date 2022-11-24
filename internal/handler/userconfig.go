package handler

import (
	"errors"

	"github.com/ktrysmt/go-bitbucket"
	"gopkg.in/yaml.v3"
)

var (
	// ErrGettingUserConfig is returned when getting the user config fails.
	ErrGettingUserConfig = errors.New("error getting user config")
	// ErrParsingUserConfig is returned when parsing the user config fails.
	ErrParsingUserConfig = errors.New("error parsing user config")
)

// UserConfig represents a repository-level configuration for SemanticPullRequests.
type UserConfig struct {
	Enabled            *bool     `yaml:"enabled"`
	TitleOnly          *bool     `yaml:"titleOnly"`
	CommitsOnly        *bool     `yaml:"commitsOnly"`
	TitleAndCommits    *bool     `yaml:"titleAndCommits"`
	AnyCommit          *bool     `yaml:"anyCommit"`
	Scopes             *[]string `yaml:"scopes"`
	Types              *[]string `yaml:"types"`
	AllowMergeCommits  *bool     `yaml:"allowMergeCommits"`
	AllowRevertCommits *bool     `yaml:"allowRevertCommits"`
}

// DefaultUserConfig returns an initialised UserConfig with default values.
func DefaultUserConfig() *UserConfig {
	enabled := true
	titleOnly := false
	commitsOnly := false
	titleAndCommits := false
	anyCommits := false
	allowMergeCommits := false
	allowRevertCommits := false

	return &UserConfig{
		Enabled:            &enabled,
		TitleOnly:          &titleOnly,
		CommitsOnly:        &commitsOnly,
		TitleAndCommits:    &titleAndCommits,
		AnyCommit:          &anyCommits,
		AllowMergeCommits:  &allowMergeCommits,
		AllowRevertCommits: &allowRevertCommits,
	}
}

// GetUserConfig returns the user config for a given repository.
func GetUserConfig(client *bitbucket.Client, owner, repoSlug string) (*UserConfig, error) {
	userConfig := DefaultUserConfig()

	blob, err := client.Repositories.Repository.GetFileBlob(&bitbucket.RepositoryBlobOptions{
		Owner:    owner,
		RepoSlug: repoSlug,
		Ref:      "HEAD",
		Path:     ".bitbucket/semantic.yml",
	})
	if err != nil {
		return userConfig, ErrGettingUserConfig
	}

	if blob != nil {
		if err := yaml.Unmarshal(blob.Content, userConfig); err != nil {
			return userConfig, ErrParsingUserConfig
		}
	}

	return userConfig, nil
}

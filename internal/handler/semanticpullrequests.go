package handler

import (
	"errors"
	"fmt"
	"strings"

	webhook "github.com/go-playground/webhooks/v6/bitbucket"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"
	"go.uber.org/zap"
)

// SemanticPullRequests validates Bitbucket pull-requests.
type SemanticPullRequests struct {
	Client       *bitbucket.Client
	Conventional conventionalcommits.Machine
	FreeForm     conventionalcommits.Machine
	Hook         *webhook.Webhook
	Logger       *zap.Logger
}

var errParsingCommits = errors.New("error parsing commits")

// NewSemanticPullRequests returns an initialized SemanticPullRequests.
func NewSemanticPullRequests(username, password string, logger *zap.Logger) (*SemanticPullRequests, error) {
	client := bitbucket.NewBasicAuth(username, password)

	conventional := parser.NewMachine(
		conventionalcommits.WithTypes(conventionalcommits.TypesConventional),
	)

	freeForm := parser.NewMachine(
		conventionalcommits.WithTypes(conventionalcommits.TypesFreeForm),
	)

	hook, err := webhook.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize webhook: %w", err)
	}

	return &SemanticPullRequests{
		Client:       client,
		Conventional: conventional,
		FreeForm:     freeForm,
		Hook:         hook,
		Logger:       logger,
	}, nil
}

// IsSemanticMessage validates the semantic of a given message.
func (spr *SemanticPullRequests) IsSemanticMessage(cfg *UserConfig, msg string) bool {
	if *cfg.AllowMergeCommits && strings.HasPrefix(msg, "Merge") {
		return true
	}

	if *cfg.AllowRevertCommits && strings.HasPrefix(msg, "Revert") {
		return true
	}

	var ccMsg conventionalcommits.Message
	var err error
	if cfg.Types == nil {
		ccMsg, err = spr.Conventional.Parse([]byte(msg))
	} else {
		ccMsg, err = spr.FreeForm.Parse([]byte(msg))
	}

	if err != nil {
		spr.Logger.Debug(
			"failed to parse message",
			zap.String("message", msg),
			zap.Error(err),
		)

		return false
	}

	cc, ok := ccMsg.(*conventionalcommits.ConventionalCommit)
	if !ok {
		spr.Logger.Debug("failed to convert parsed message to conventional commit")

		return false
	}

	isScopeValid := true
	if cfg.Scopes != nil && cc.Scope != nil {
		for _, s := range strings.Split(*cc.Scope, ",") {
			if !Contains(*cfg.Scopes, strings.TrimSpace(s)) {
				isScopeValid = false
			}
		}
	}

	isTypeValid := cfg.Types == nil || Contains(*cfg.Types, cc.Type)

	return isScopeValid && isTypeValid
}

func (spr *SemanticPullRequests) areSemanticCommits(cfg *UserConfig, commits []interface{}, anyCommit bool) bool {
	var c map[string]interface{}

	var isSemantic, ok bool

	for _, ciface := range commits {
		c, ok = ciface.(map[string]interface{})
		if !ok {
			spr.Logger.Error("failed to convert commit type")
		}

		msg, ok := c["message"].(string)
		if !ok {
			spr.Logger.Error("failed to convert commit message type")
		}

		// ¯\_(ツ)_/¯ Bitbucket Cloud adds a trailing newline to messages.
		// If there is a single newline after the description, but no body,
		// the parser fails with conventionalcommits.ErrMissingBlankLineAtBeginning
		// The Git CLI trims trailing white-spaces when committing,
		// so it is definitely not coming from the user
		msg = strings.TrimSuffix(msg, "\n")

		isSemantic = spr.IsSemanticMessage(cfg, msg)
		if anyCommit && isSemantic {
			return true
		} else if !anyCommit && !isSemantic {
			return false
		}
	}

	return !anyCommit
}

func (spr *SemanticPullRequests) getCommits(owner, repoSlug, prID string) ([]interface{}, error) {
	resIface, err := spr.Client.Repositories.PullRequests.Commits(&bitbucket.PullRequestsOptions{
		Owner:    owner,
		RepoSlug: repoSlug,
		ID:       prID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	result, ok := resIface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w when converting result type: %+v", errParsingCommits, resIface)
	}

	commits, ok := result["values"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w when converting values type: %+v", errParsingCommits, resIface)
	}

	return commits, nil
}

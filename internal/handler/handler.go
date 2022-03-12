package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	webhook "github.com/go-playground/webhooks/v6/bitbucket"
	"github.com/ktrysmt/go-bitbucket"
	"go.uber.org/zap"
)

type contextKey int

const (
	// SemanticPullRequestsKey is the http.Request context key containing the initialized SemanticPullRequests.
	SemanticPullRequestsKey contextKey = 0
)

func isSemanticPullRequest(cfg *UserConfig, hasSemanticTitle, hasSemanticCommits bool) bool {
	switch {
	case !*cfg.Enabled:
		return true
	case *cfg.TitleOnly:
		return hasSemanticTitle
	case *cfg.CommitsOnly:
		return hasSemanticCommits
	case *cfg.TitleAndCommits:
		return hasSemanticTitle && hasSemanticCommits
	default:
		return hasSemanticTitle || hasSemanticCommits
	}
}

func getStatusDescription(cfg *UserConfig, hasSemanticTitle, hasSemanticCommits, isSemantic bool) string {
	switch {
	case !*cfg.Enabled:
		return "skipped; check disabled in semantic.yml config"
	case isSemantic && *cfg.TitleAndCommits:
		return "ready to be merged, squashed or fast-forwarded"
	case !isSemantic && *cfg.TitleAndCommits:
		return "add a semantic commit AND PR title"
	case hasSemanticTitle && !*cfg.CommitsOnly:
		return "ready to be squashed"
	case hasSemanticCommits && !*cfg.TitleOnly:
		return "ready to be merged or fast-forwarded"
	case *cfg.TitleOnly:
		return "add a semantic PR title"
	case *cfg.CommitsOnly && *cfg.AnyCommit:
		return "add a semantic commit"
	case *cfg.CommitsOnly:
		return "make sure every commit is semantic"
	default:
		return "add a semantic commit or PR title"
	}
}

// HandlePullRequestUpdate handles pull-request update events.
func HandlePullRequestUpdate(w http.ResponseWriter, r *http.Request) {
	// Useful for simple heath check
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return
	}

	spr, ok := r.Context().Value(SemanticPullRequestsKey).(*SemanticPullRequests)
	if !ok {
		log.Println("Error: failed to retrieve semanticPullRequests from context")
	}

	pl, err := spr.Hook.Parse(r, webhook.PullRequestCreatedEvent, webhook.PullRequestUpdatedEvent)
	if err != nil {
		spr.Logger.Error("failed to parse request", zap.Error(err))

		switch {
		case errors.Is(err, webhook.ErrEventNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, webhook.ErrEventNotSpecifiedToParse):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, webhook.ErrInvalidHTTPMethod):
			w.WriteHeader(http.StatusMethodNotAllowed)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	var payload webhook.PullRequestUpdatedPayload

	switch pl := pl.(type) {
	case webhook.PullRequestUpdatedPayload:
		payload = pl
	case webhook.PullRequestCreatedPayload:
		payload = webhook.PullRequestUpdatedPayload(pl)
	default:
		spr.Logger.Error("failed to convert webhook payload type", zap.Reflect("payload", pl))

		return
	}

	logReqFields := []zap.Field{
		zap.String("repository", payload.Repository.FullName),
		zap.Int64("pull_request_id", payload.PullRequest.ID),
		zap.String("revision", payload.PullRequest.Source.Commit.Hash),
	}
	spr.Logger.Info(
		"handling pull request update",
		logReqFields...,
	)

	userConfig, err := GetUserConfig(spr.Client, payload.Repository.Owner.UUID, payload.Repository.UUID)
	if err != nil {
		spr.Logger.Debug(
			"failed to get user config",
			append(logReqFields, zap.Error(err))...,
		)
	}

	hasSemanticTitle := spr.IsSemanticMessage(userConfig, payload.PullRequest.Title)

	commits, err := spr.getCommits(
		payload.Repository.Owner.UUID,
		payload.Repository.UUID,
		strconv.Itoa(int(payload.PullRequest.ID)),
	)
	if err != nil {
		spr.Logger.Error(
			"failed to get commits",
			append(logReqFields, zap.Error(err))...,
		)
	}

	var hasSemanticCommits bool
	if commits != nil {
		hasSemanticCommits = spr.AreSemanticCommits(userConfig, commits)
	} else {
		hasSemanticCommits = false
	}

	isSemantic := isSemanticPullRequest(userConfig, hasSemanticTitle, hasSemanticCommits)

	var state string
	if isSemantic {
		state = "SUCCESSFUL"
	} else {
		state = "FAILED"
	}

	description := getStatusDescription(userConfig, hasSemanticTitle, hasSemanticCommits, isSemantic)

	cso := &bitbucket.CommitStatusOptions{
		Key:         "Semantic Pull Request",
		State:       state,
		Description: description,
		Url:         "https://github.com/maxbrunet/bitbucket-semantic-pull-requests",
	}

	co := &bitbucket.CommitsOptions{
		Owner:    payload.Repository.Owner.UUID,
		RepoSlug: payload.Repository.UUID,
		Revision: payload.PullRequest.Source.Commit.Hash,
	}

	if _, err := spr.Client.Repositories.Commits.CreateCommitStatus(co, cso); err != nil {
		spr.Logger.Error(
			"failed to create commit status",
			append(logReqFields, zap.Error(err))...,
		)
	}
}

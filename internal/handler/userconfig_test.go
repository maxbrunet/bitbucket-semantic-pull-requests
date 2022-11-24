package handler_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"

	"github.com/maxbrunet/bitbucket-semantic-pull-requests/internal/handler"
)

func TestUserConfigNotFound(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get(fmt.Sprintf(
			"/2.0/repositories/%s/%s/src/HEAD/.bitbucket/semantic.yml",
			ownerUUID, repositoryUUID,
		)).
		MatchHeader("Authorization", authorizationHeader).
		Reply(404)

	cfg, err := handler.GetUserConfig(spr.Client, ownerUUID, repositoryUUID)

	require.Equal(t, true, *cfg.Enabled)
	require.Equal(t, false, *cfg.TitleOnly)
	require.Equal(t, false, *cfg.CommitsOnly)
	require.Equal(t, false, *cfg.TitleOnly)
	require.Equal(t, false, *cfg.TitleAndCommits)
	require.Equal(t, false, *cfg.AnyCommit)
	require.Nil(t, cfg.Scopes)
	require.Nil(t, cfg.Types)
	require.Equal(t, false, *cfg.AllowMergeCommits)
	require.Equal(t, false, *cfg.AllowRevertCommits)
	require.Equal(t, handler.ErrGettingUserConfig, err)
	require.Equal(t, true, gock.IsDone())
}

func TestUserConfigInvalid(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get(fmt.Sprintf(
			"/2.0/repositories/%s/%s/src/HEAD/.bitbucket/semantic.yml",
			ownerUUID, repositoryUUID,
		)).
		MatchHeader("Authorization", authorizationHeader).
		Reply(200).
		Body(strings.NewReader("invalid config"))

	cfg, err := handler.GetUserConfig(spr.Client, ownerUUID, repositoryUUID)

	require.Equal(t, true, *cfg.Enabled)
	require.Equal(t, false, *cfg.TitleOnly)
	require.Equal(t, false, *cfg.CommitsOnly)
	require.Equal(t, false, *cfg.TitleOnly)
	require.Equal(t, false, *cfg.TitleAndCommits)
	require.Equal(t, false, *cfg.AnyCommit)
	require.Nil(t, cfg.Scopes)
	require.Nil(t, cfg.Types)
	require.Equal(t, false, *cfg.AllowMergeCommits)
	require.Equal(t, false, *cfg.AllowRevertCommits)
	require.Equal(t, handler.ErrParsingUserConfig, err)
	require.Equal(t, true, gock.IsDone())
}

func TestUserConfigFullValid(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get(fmt.Sprintf(
			"/2.0/repositories/%s/%s/src/HEAD/.bitbucket/semantic.yml",
			ownerUUID, repositoryUUID,
		)).
		MatchHeader("Authorization", authorizationHeader).
		Reply(200).
		BodyString(`enabled: false
titleOnly: true
commitsOnly: true
titleAndCommits: true
anyCommit: true
scopes:
- scope1
- scope2
types:
- type1
- type2
allowMergeCommits: true
allowRevertCommits: true
`)

	cfg, err := handler.GetUserConfig(spr.Client, ownerUUID, repositoryUUID)

	require.Equal(t, false, *cfg.Enabled)
	require.Equal(t, true, *cfg.TitleOnly)
	require.Equal(t, true, *cfg.CommitsOnly)
	require.Equal(t, true, *cfg.TitleOnly)
	require.Equal(t, true, *cfg.TitleAndCommits)
	require.Equal(t, true, *cfg.AnyCommit)
	require.Equal(t, []string{"scope1", "scope2"}, *cfg.Scopes)
	require.Equal(t, []string{"type1", "type2"}, *cfg.Types)
	require.Equal(t, true, *cfg.AllowMergeCommits)
	require.Equal(t, true, *cfg.AllowRevertCommits)
	require.Nil(t, err)
	require.Equal(t, true, gock.IsDone())
}

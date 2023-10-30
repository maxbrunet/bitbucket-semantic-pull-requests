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

	require.True(t, *cfg.Enabled)
	require.False(t, *cfg.TitleOnly)
	require.False(t, *cfg.CommitsOnly)
	require.False(t, *cfg.TitleOnly)
	require.False(t, *cfg.TitleAndCommits)
	require.False(t, *cfg.AnyCommit)
	require.Nil(t, cfg.Scopes)
	require.Nil(t, cfg.Types)
	require.False(t, *cfg.AllowMergeCommits)
	require.False(t, *cfg.AllowRevertCommits)
	require.Equal(t, handler.ErrGettingUserConfig, err)
	require.True(t, gock.IsDone())
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

	require.True(t, *cfg.Enabled)
	require.False(t, *cfg.TitleOnly)
	require.False(t, *cfg.CommitsOnly)
	require.False(t, *cfg.TitleOnly)
	require.False(t, *cfg.TitleAndCommits)
	require.False(t, *cfg.AnyCommit)
	require.Nil(t, cfg.Scopes)
	require.Nil(t, cfg.Types)
	require.False(t, *cfg.AllowMergeCommits)
	require.False(t, *cfg.AllowRevertCommits)
	require.Equal(t, handler.ErrParsingUserConfig, err)
	require.True(t, gock.IsDone())
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

	require.False(t, *cfg.Enabled)
	require.True(t, *cfg.TitleOnly)
	require.True(t, *cfg.CommitsOnly)
	require.True(t, *cfg.TitleOnly)
	require.True(t, *cfg.TitleAndCommits)
	require.True(t, *cfg.AnyCommit)
	require.Equal(t, []string{"scope1", "scope2"}, *cfg.Scopes)
	require.Equal(t, []string{"type1", "type2"}, *cfg.Types)
	require.True(t, *cfg.AllowMergeCommits)
	require.True(t, *cfg.AllowRevertCommits)
	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

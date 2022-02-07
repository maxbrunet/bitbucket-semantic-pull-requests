package handler_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/maxbrunet/bitbucket-semantic-pull-requests/internal/handler"
)

const (
	bitbucketUsername string = "username"
	bitbucketPassword string = "password"
)

var spr *handler.SemanticPullRequests

func init() {
	spr, _ = handler.NewSemanticPullRequests(bitbucketUsername, bitbucketPassword, zap.NewNop())
}

func TestIsSemanticMessage(t *testing.T) {
	require.Equal(t,
		true,
		spr.IsSemanticMessage(handler.DefaultUserConfig(), "fix: something"),
	)
}

func TestIsSemanticMessageWithScope(t *testing.T) {
	require.Equal(t,
		true,
		spr.IsSemanticMessage(handler.DefaultUserConfig(), "fix(subsystem): something"),
	)
}

func TestIsNotSemanticMessage(t *testing.T) {
	require.Equal(t,
		false,
		spr.IsSemanticMessage(handler.DefaultUserConfig(), "unsemantic commit message"),
	)
}

func TestIsSemanticMessageWithRestrictedScopes(t *testing.T) {
	userConfig := handler.DefaultUserConfig()

	cases := []struct {
		name     string
		message  string
		scopes   []string
		expected bool
	}{
		{
			name:     "validScope",
			message:  "fix(validScope): something",
			scopes:   []string{"validScope"},
			expected: true,
		},
		{
			name:     "invalidScope",
			message:  "fix(invalidScope): something",
			scopes:   []string{"validScope"},
			expected: false,
		},
		{
			name:     "multipleValidScopes",
			message:  "fix(validScope,anotherValidScope): something",
			scopes:   []string{"validScope", "anotherValidscope"},
			expected: true,
		},
		{
			name:     "multipleValidScopesWithSpace",
			message:  "fix(validScope, spaceAndAnotherValidScope): something",
			scopes:   []string{"validScope", "spaceAndAnotherValidscope"},
			expected: true,
		},
		{
			name:     "multipleScopesWithOneInvalid",
			message:  "fix(validScope, invalidScope): something",
			scopes:   []string{"validScope"},
			expected: false,
		},
		{
			name:     "noScope",
			message:  "fix: something",
			scopes:   []string{"validScope"},
			expected: true,
		},
		// Differs from zeke/semantic-pull-requests
		// Empty and no scope are treated the same by go-conventionalcommits (nil)
		{
			name:     "emptyScope",
			message:  "fix(): something",
			scopes:   []string{"validScope"},
			expected: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			userConfig.Scopes = &tc.scopes
			require.Equal(t,
				tc.expected,
				spr.IsSemanticMessage(userConfig, tc.message),
			)
		})
	}
}

func TestIsSemanticMessageWithAllowMergeCommits(t *testing.T) {
	userConfig := handler.DefaultUserConfig()
	allowMergeCommits := true
	userConfig.AllowMergeCommits = &allowMergeCommits

	require.Equal(t,
		true,
		spr.IsSemanticMessage(userConfig, "Merge branch 'master' into patch-1"),
	)

	scopes := []string{"scope1"}
	userConfig.Scopes = &scopes
	require.Equal(t,
		true,
		spr.IsSemanticMessage(userConfig, "Merge refs/heads/master into angry-burritos/US-335"),
	)
}

func TestIsSemanticMessageWithAllowRevertCommits(t *testing.T) {
	userConfig := handler.DefaultUserConfig()
	allowRevertCommits := true
	userConfig.AllowRevertCommits = &allowRevertCommits

	require.Equal(t,
		true,
		spr.IsSemanticMessage(userConfig, "Revert \"feat: ride unicorns\"\n"),
	)

	scopes := []string{"scope1"}
	userConfig.Scopes = &scopes
	require.Equal(t,
		true,
		spr.IsSemanticMessage(userConfig, "Revert \"feat: ride unicorns\"\n"),
	)
}

func TestIsSemanticMessageWithValidTypes(t *testing.T) {
	types := []string{
		"feat",
		"fix",
		"docs",
		"style",
		"refactor",
		"perf",
		"test",
		"build",
		"ci",
		"chore",
		"revert",
	}

	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			require.Equal(t,
				true,
				spr.IsSemanticMessage(handler.DefaultUserConfig(), typ+": something"),
			)
		})
	}
}

func TestIsNotSemanticMessageWithInvalidType(t *testing.T) {
	require.Equal(t,
		false,
		spr.IsSemanticMessage(handler.DefaultUserConfig(), "alternative: something"),
	)
}

func TestIsSemanticMessageWithValidCustomTypes(t *testing.T) {
	userConfig := handler.DefaultUserConfig()
	customTypes := []string{
		"alternative",
		"improvement",
	}
	userConfig.Types = &customTypes

	types := []string{
		"alternative",
		"improvement",
	}

	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			require.Equal(t,
				true,
				spr.IsSemanticMessage(userConfig, typ+": something"),
			)
		})
	}
}

func TestIsSemanticMessageWithInvalidCustomTypes(t *testing.T) {
	userConfig := handler.DefaultUserConfig()
	customTypes := []string{
		"alternative",
		"improvement",
	}
	userConfig.Types = &customTypes

	types := []string{
		"feat",
		"fix",
		"docs",
		"style",
		"refactor",
		"perf",
		"test",
		"build",
		"ci",
		"chore",
		"revert",
	}

	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			require.Equal(t,
				false,
				spr.IsSemanticMessage(userConfig, typ+": something"),
			)
		})
	}
}

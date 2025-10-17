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
	t.Parallel()

	require.True(
		t,
		spr.IsSemanticMessage(
			handler.NewSemanticMachine(true),
			handler.DefaultUserConfig(),
			"fix: something",
		),
	)
}

func TestIsSemanticMessageWithScope(t *testing.T) {
	t.Parallel()

	require.True(
		t,
		spr.IsSemanticMessage(
			handler.NewSemanticMachine(true),
			handler.DefaultUserConfig(),
			"fix(subsystem): something",
		),
	)
}

func TestIsNotSemanticMessage(t *testing.T) {
	t.Parallel()

	require.False(
		t,
		spr.IsSemanticMessage(
			handler.NewSemanticMachine(true),
			handler.DefaultUserConfig(),
			"unsemantic commit message",
		),
	)
}

func TestIsSemanticMessageWithRestrictedScopes(t *testing.T) {
	t.Parallel()

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

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userConfig := handler.DefaultUserConfig()
			userConfig.Scopes = &tc.scopes

			require.Equal(t,
				tc.expected,
				spr.IsSemanticMessage(handler.NewSemanticMachine(true), userConfig, tc.message),
			)
		})
	}
}

func TestIsSemanticMessageWithAllowMergeCommits(t *testing.T) {
	t.Parallel()

	machine := handler.NewSemanticMachine(true)
	userConfig := handler.DefaultUserConfig()
	allowMergeCommits := true
	userConfig.AllowMergeCommits = &allowMergeCommits

	require.True(t,
		spr.IsSemanticMessage(machine, userConfig, "Merge branch 'master' into patch-1"),
	)

	scopes := []string{"scope1"}
	userConfig.Scopes = &scopes
	require.True(
		t,
		spr.IsSemanticMessage(
			machine,
			userConfig,
			"Merge refs/heads/master into angry-burritos/US-335",
		),
	)
}

func TestIsSemanticMessageWithAllowRevertCommits(t *testing.T) {
	t.Parallel()

	machine := handler.NewSemanticMachine(true)
	userConfig := handler.DefaultUserConfig()
	allowRevertCommits := true
	userConfig.AllowRevertCommits = &allowRevertCommits

	require.True(t,
		spr.IsSemanticMessage(machine, userConfig, "Revert \"feat: ride unicorns\"\n"),
	)

	scopes := []string{"scope1"}
	userConfig.Scopes = &scopes
	require.True(t,
		spr.IsSemanticMessage(machine, userConfig, "Revert \"feat: ride unicorns\"\n"),
	)
}

func TestIsSemanticMessageWithValidTypes(t *testing.T) {
	t.Parallel()

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

	for _, tt := range types {
		tc := tt
		t.Run(tc, func(t *testing.T) {
			t.Parallel()

			require.True(
				t,
				spr.IsSemanticMessage(
					handler.NewSemanticMachine(true),
					handler.DefaultUserConfig(),
					tc+": something",
				),
			)
		})
	}
}

func TestIsNotSemanticMessageWithInvalidType(t *testing.T) {
	t.Parallel()

	require.False(
		t,
		spr.IsSemanticMessage(
			handler.NewSemanticMachine(true),
			handler.DefaultUserConfig(),
			"alternative: something",
		),
	)
}

func TestIsSemanticMessageWithValidCustomTypes(t *testing.T) {
	t.Parallel()

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

	for _, tt := range types {
		tc := tt
		t.Run(tc, func(t *testing.T) {
			t.Parallel()

			require.True(
				t,
				spr.IsSemanticMessage(
					handler.NewSemanticMachine(false),
					userConfig,
					tc+": something",
				),
			)
		})
	}
}

func TestIsSemanticMessageWithInvalidCustomTypes(t *testing.T) {
	t.Parallel()

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

	for _, tt := range types {
		tc := tt
		t.Run(tc, func(t *testing.T) {
			t.Parallel()

			require.False(
				t,
				spr.IsSemanticMessage(
					handler.NewSemanticMachine(true),
					userConfig,
					tc+": something",
				),
			)
		})
	}
}

func TestAreSemanticCommits(t *testing.T) {
	t.Parallel()

	defaultCfg := handler.DefaultUserConfig()

	anyCommitCfg := handler.DefaultUserConfig()
	anyCommit := true
	anyCommitCfg.AnyCommit = &anyCommit

	valid := []any{
		map[string]any{
			"message": "feat: potato\n",
		},
	}

	partiallyValid := []any{
		map[string]any{
			"message": "feat: banana\n",
		},
		map[string]any{
			"message": "unicorn\n",
		},
		map[string]any{
			"message": "feat: potato\n",
		},
	}

	malformed := []any{"not a commit"}

	partiallyMalformed := []any{
		map[string]any{
			"message": "feat: banana\n",
		},
		map[string]any{
			"unknown": "not a commit",
		},
		map[string]any{
			"message": "feat: potato\n",
		},
	}

	cases := []struct {
		name     string
		cfg      *handler.UserConfig
		commits  []any
		expected bool
	}{
		{
			name:     "DefaultConfig/Valid",
			cfg:      defaultCfg,
			commits:  valid,
			expected: true,
		},
		{
			name:     "AnyCommitConfig/Valid",
			cfg:      anyCommitCfg,
			commits:  valid,
			expected: true,
		},
		{
			name:     "DefaultConfig/PartiallyValid",
			cfg:      defaultCfg,
			commits:  partiallyValid,
			expected: false,
		},
		{
			name:     "AnyCommitConfig/PartiallyValid",
			cfg:      anyCommitCfg,
			commits:  partiallyValid,
			expected: true,
		},
		{
			name:     "DefaultConfig/Malformed",
			cfg:      defaultCfg,
			commits:  malformed,
			expected: false,
		},
		{
			name:     "AnyCommitConfig/Malformed",
			cfg:      anyCommitCfg,
			commits:  malformed,
			expected: false,
		},
		{
			name:     "DefaultConfig/PartiallyMalformed",
			cfg:      defaultCfg,
			commits:  partiallyMalformed,
			expected: false,
		},
		{
			name:     "AnyCommitConfig/PartiallyMalformed",
			cfg:      anyCommitCfg,
			commits:  partiallyMalformed,
			expected: true,
		},
	}

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t,
				tc.expected,
				spr.AreSemanticCommits(handler.NewSemanticMachine(true), tc.cfg, tc.commits),
			)
		})
	}
}

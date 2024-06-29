package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
)

const (
	authorizationHeader string = "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
	ownerUUID           string = "{bf52b9af-d72c-48b7-8258-7c64f3788fe5}"
	repositoryUUID      string = "{e8a4e930-9104-449e-855d-ee814b23a36b}"
	commitHash          string = "e10dae226959c2194f2b07b077c07762d93821cf"
	pullRequestID       int    = 123
)

type testCase struct {
	name                string
	event               string
	prTitle             string
	commits             []interface{}
	config              string
	expectedState       string
	expectedDescription string
}

func unsemanticCommits() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"message": "fix something\n",
		},
		map[string]interface{}{
			"message": "fix something else\n",
		},
	}
}

func semanticCommits() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"message": "build(scope1): something\n",
		},
		map[string]interface{}{
			"message": "build(scope2): something else\n",
		},
	}
}

func runTestCase(t *testing.T, tc *testCase) {
	t.Helper()
	defer gock.CleanUnmatchedRequest()
	defer gock.Off()

	data, err := json.Marshal(
		map[string]interface{}{
			"pullrequest": map[string]interface{}{
				"id":    pullRequestID,
				"title": tc.prTitle,
				"source": map[string]interface{}{
					"commit": map[string]interface{}{
						"hash": commitHash,
					},
				},
			},
			"repository": map[string]interface{}{
				"owner": map[string]interface{}{
					"uuid": ownerUUID,
				},
				"uuid": repositoryUUID,
			},
		},
	)
	if err != nil {
		t.Error("fail to prepare mock payload")
	}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
	req.Header.Set("X-Event-Key", tc.event)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	gock.New("https://api.bitbucket.org").
		Get(fmt.Sprintf(
			"/2.0/repositories/%s/%s/src/HEAD/.bitbucket/semantic.yml",
			ownerUUID, repositoryUUID,
		)).
		MatchHeader("Authorization", authorizationHeader).
		Reply(200).
		Body(strings.NewReader(tc.config))

	gock.New("https://api.bitbucket.org").
		Get(fmt.Sprintf(
			"/2.0/repositories/%s/%s/pullrequests/%d/commits",
			ownerUUID, repositoryUUID, pullRequestID,
		)).
		MatchHeader("Authorization", authorizationHeader).
		Reply(200).
		JSON(map[string]interface{}{
			"values": tc.commits,
		})

	gock.New("https://api.bitbucket.org").
		Post(fmt.Sprintf(
			"/2.0/repositories/%s/%s/commit/%s/status",
			ownerUUID, repositoryUUID, commitHash,
		)).
		MatchHeader("Authorization", authorizationHeader).
		MatchHeader("Content-Type", "application/json").
		JSON(map[string]interface{}{
			"name":        "",
			"key":         "Semantic Pull Request",
			"state":       tc.expectedState,
			"description": tc.expectedDescription,
			"url":         "https://github.com/maxbrunet/bitbucket-semantic-pull-requests",
		}).
		Reply(200).
		JSON(map[string]string{})

	spr.HandlePullRequestUpdate(rec, req)

	if gock.HasUnmatchedRequest() {
		t.Error("Cound not matched all requests:")
		for _, ureq := range gock.GetUnmatchedRequests() {
			bytes, _ := httputil.DumpRequestOut(ureq, true)
			t.Log(string(bytes))
		}
	}
	require.True(t, gock.IsDone())
}

func TestGetRoot(t *testing.T) {
	t.Parallel()

	for _, m := range []string{http.MethodGet, http.MethodHead} {
		t.Run(m, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(m, "/", nil)
			rec := httptest.NewRecorder()

			spr.HandlePullRequestUpdate(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			require.Equal(t, 200, res.StatusCode)
		})
	}
}

func TestVanillaConfig(t *testing.T) {
	cases := []testCase{
		{
			name:                "FailedWithNoSemanticCommitsAndNoSemanticTitle",
			event:               "pullrequest:created",
			prTitle:             "do a thing",
			commits:             unsemanticCommits(),
			config:              "",
			expectedState:       "FAILED",
			expectedDescription: "add a semantic commit or PR title",
		},
		{
			name:                "SuccessfulWithSemanticCommits",
			event:               "pullrequest:updated",
			prTitle:             "bananas",
			commits:             semanticCommits(),
			config:              "",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be merged or fast-forwarded",
		},
		{
			name:                "SuccessfulWithSemanticTitle",
			event:               "pullrequest:updated",
			prTitle:             "build: do a thing",
			commits:             unsemanticCommits(),
			config:              "",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be squashed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, &tc)
		})
	}
}

func TestNotEnabledSuccessful(t *testing.T) {
	tc := testCase{
		event:               "pullrequest:updated",
		prTitle:             "do a thing",
		commits:             unsemanticCommits(),
		config:              "enabled: false",
		expectedState:       "SUCCESSFUL",
		expectedDescription: "skipped; check disabled in semantic.yml config",
	}

	runTestCase(t, &tc)
}

func TestScopes(t *testing.T) {
	cases := []testCase{
		{
			name:                "FailedWithSemanticTitleAndInvalidScope",
			event:               "pullrequest:updated",
			prTitle:             "fix(scope3): do a thing",
			commits:             semanticCommits(),
			config:              "titleOnly: true\nscopes: [scope1, scope2]",
			expectedState:       "FAILED",
			expectedDescription: "add a semantic PR title",
		},
		{
			name:                "SuccessfulWithSemanticTitleAndValidScope",
			event:               "pullrequest:updated",
			prTitle:             "fix(scope1): bananas",
			commits:             semanticCommits(),
			config:              "titleOnly: true\nscopes: [scope1, scope2]",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be squashed",
		},
		{
			name:                "FailedWithSemanticCommitsAndInvalidScopes",
			event:               "pullrequest:updated",
			prTitle:             "fix(scope3): do a thing",
			commits:             semanticCommits(),
			config:              "commitsOnly: true\nscopes: [scope3, scope4]",
			expectedState:       "FAILED",
			expectedDescription: "make sure every commit is semantic",
		},
		{
			name:                "SuccessfulWithSemanticCommitsAndValidScopes",
			event:               "pullrequest:updated",
			prTitle:             "fix(scope1): bananas",
			commits:             semanticCommits(),
			config:              "commitsOnly: true\nscopes: [scope1, scope2]",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be merged or fast-forwarded",
		},
		{
			name:                "SuccessfulWithSemanticTitleAndNoScope",
			event:               "pullrequest:updated",
			prTitle:             "fix: bananas",
			commits:             semanticCommits(),
			config:              "titleOnly: true\nscopes: [scope1]",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be squashed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, &tc)
		})
	}
}

func TestTypes(t *testing.T) {
	cases := []testCase{
		{
			name:                "FailedWithSemanticTitleAndInvalidType",
			event:               "pullrequest:updated",
			prTitle:             "fix(scope3): do a thing",
			commits:             semanticCommits(),
			config:              "titleOnly: true\ntypes: [type1, type2]",
			expectedState:       "FAILED",
			expectedDescription: "add a semantic PR title",
		},
		{
			name:                "SuccessfulWithSemanticTitleAndValidType",
			event:               "pullrequest:updated",
			prTitle:             "type1(scope1): bananas",
			commits:             semanticCommits(),
			config:              "titleOnly: true\ntypes: [type1, type2]",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be squashed",
		},
		{
			name:                "FailedWithSemanticCommitsAndInvalidTypes",
			event:               "pullrequest:updated",
			prTitle:             "fix(scope3): do a thing",
			commits:             semanticCommits(),
			config:              "commitsOnly: true\ntypes: [type1, type2]",
			expectedState:       "FAILED",
			expectedDescription: "make sure every commit is semantic",
		},
		{
			name:    "SuccessfulWithSemanticCommitsAndValidTypes",
			event:   "pullrequest:updated",
			prTitle: "fix(scope1): bananas",
			commits: []interface{}{
				map[string]interface{}{
					"message": "type1(scope1): something\n",
				},
				map[string]interface{}{
					"message": "type2(scope2): something else\n",
				},
			},
			config:              "commitsOnly: true\ntypes: [type1, type2]",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be merged or fast-forwarded",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, &tc)
		})
	}
}

func TestCommitsOnly(t *testing.T) {
	cases := []testCase{
		{
			name:                "FailedWithNoSemanticCommits",
			event:               "pullrequest:updated",
			prTitle:             "do a thing",
			commits:             unsemanticCommits(),
			config:              "commitsOnly: true",
			expectedState:       "FAILED",
			expectedDescription: "make sure every commit is semantic",
		},
		{
			name:                "FailedWithNoSemanticCommitsButSemanticTitle",
			event:               "pullrequest:updated",
			prTitle:             "fix: do a thing",
			commits:             unsemanticCommits(),
			config:              "commitsOnly: true",
			expectedState:       "FAILED",
			expectedDescription: "make sure every commit is semantic",
		},
		{
			name:                "FailedWithSomeSemanticCommitsButNotAll",
			event:               "pullrequest:updated",
			prTitle:             "fix: do a thing",
			commits:             append(unsemanticCommits(), semanticCommits()...),
			config:              "commitsOnly: true",
			expectedState:       "FAILED",
			expectedDescription: "make sure every commit is semantic",
		},
		{
			name:                "SuccessfulWithSemanticCommits",
			event:               "pullrequest:updated",
			prTitle:             "bananas",
			commits:             semanticCommits(),
			config:              "commitsOnly: true",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be merged or fast-forwarded",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, &tc)
		})
	}
}

func TestTitleOnly(t *testing.T) {
	cases := []testCase{
		{
			name:                "FailedWithNoSemanticTitle",
			event:               "pullrequest:updated",
			prTitle:             "do a thing",
			commits:             unsemanticCommits(),
			config:              "titleOnly: true",
			expectedState:       "FAILED",
			expectedDescription: "add a semantic PR title",
		},
		{
			name:                "FailedWithSemanticCommitsAndNoSemanticTitle",
			event:               "pullrequest:updated",
			prTitle:             "do a thing",
			commits:             semanticCommits(),
			config:              "titleOnly: true",
			expectedState:       "FAILED",
			expectedDescription: "add a semantic PR title",
		},
		{
			name:                "SuccessfulWithSemanticTitle",
			event:               "pullrequest:updated",
			prTitle:             "build: do a thing",
			commits:             unsemanticCommits(),
			config:              "titleOnly: true",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be squashed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, &tc)
		})
	}
}

func TestTitleAndCommits(t *testing.T) {
	cases := []testCase{
		{
			name:                "FailedWithNoSemanticTitle",
			event:               "pullrequest:updated",
			prTitle:             "do a thing",
			commits:             unsemanticCommits(),
			config:              "titleAndCommits: true",
			expectedState:       "FAILED",
			expectedDescription: "add a semantic commit AND PR title",
		},
		{
			name:                "FailedWithNoSemanticTitleButSemanticCommits",
			event:               "pullrequest:updated",
			prTitle:             "do a thing",
			commits:             semanticCommits(),
			config:              "titleAndCommits: true",
			expectedState:       "FAILED",
			expectedDescription: "add a semantic commit AND PR title",
		},
		{
			name:                "FailedWithSemanticTitleButNoSemanticCommits",
			event:               "pullrequest:updated",
			prTitle:             "chore: do a thing",
			commits:             unsemanticCommits(),
			config:              "titleAndCommits: true",
			expectedState:       "FAILED",
			expectedDescription: "add a semantic commit AND PR title",
		},
		{
			name:                "SuccessfulWithSemanticTitleAndSemanticCommits",
			event:               "pullrequest:updated",
			prTitle:             "chore: do a thing",
			commits:             semanticCommits(),
			config:              "titleAndCommits: true",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be merged, squashed or fast-forwarded",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, &tc)
		})
	}
}

func TestAnyCommits(t *testing.T) {
	type testVariant struct {
		configOption        string
		expectedDescription string
	}

	matrixes := []struct {
		name     string
		variants []testVariant
		testCase testCase
	}{
		{
			name: "FailedWith",
			variants: []testVariant{
				{
					configOption:        "commitsOnly",
					expectedDescription: "add a semantic commit",
				},
				{
					configOption:        "titleAndCommits",
					expectedDescription: "add a semantic commit AND PR title",
				},
			},
			testCase: testCase{
				name:          "AndNoSemanticCommits",
				event:         "pullrequest:created",
				prTitle:       "fix: bananas",
				commits:       unsemanticCommits(),
				config:        ": true\nanyCommit: true",
				expectedState: "FAILED",
			},
		},
		{
			name: "SuccessfulWith",
			variants: []testVariant{
				{
					configOption:        "commitsOnly",
					expectedDescription: "ready to be merged or fast-forwarded",
				},
				{
					configOption:        "titleAndCommits",
					expectedDescription: "ready to be merged, squashed or fast-forwarded",
				},
			},
			testCase: testCase{
				name:          "AndSomeSemanticCommits",
				event:         "pullrequest:created",
				prTitle:       "fix: bananas",
				commits:       append(unsemanticCommits(), semanticCommits()...),
				config:        ": true\nanyCommit: true",
				expectedState: "SUCCESSFUL",
			},
		},
	}

	for _, m := range matrixes {
		t.Run(m.name, func(t *testing.T) {
			for _, v := range m.variants {
				tc := m.testCase
				tc.config = v.configOption + tc.config
				tc.expectedDescription = v.expectedDescription
				t.Run(v.configOption+tc.name, func(t *testing.T) {
					runTestCase(t, &tc)
				})
			}
		})
	}
}

func TestAllowMergeCommits(t *testing.T) {
	cases := []testCase{
		{
			name:    "SuccessfulWithSemanticAndMergeCommitsWhenEnabled",
			event:   "pullrequest:updated",
			prTitle: "fix: bananas",
			commits: append(semanticCommits(), map[string]interface{}{
				"message": "Merge branch 'master' into feature/logout\n",
			}),
			config:              "commitsOnly: true\nallowMergeCommits: true",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be merged or fast-forwarded",
		},
		{
			name:    "FailedWithSemanticAndMergeCommitsWhenDisabled",
			event:   "pullrequest:updated",
			prTitle: "fix: bananas",
			commits: append(semanticCommits(), map[string]interface{}{
				"message": "Merge branch 'master' into feature/logout\n",
			}),
			config:              "commitsOnly: true\nallowMergeCommits: false",
			expectedState:       "FAILED",
			expectedDescription: "make sure every commit is semantic",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, &tc)
		})
	}
}

func TestAllowRevertCommits(t *testing.T) {
	cases := []testCase{
		{
			name:    "SuccessfulWithSemanticAndRevertCommitsWhenEnabled",
			event:   "pullrequest:updated",
			prTitle: "fix: bananas",
			commits: append(semanticCommits(), map[string]interface{}{
				"message": "Revert \"feat: ride unicorns\"\n",
			}),
			config:              "commitsOnly: true\nallowRevertCommits: true",
			expectedState:       "SUCCESSFUL",
			expectedDescription: "ready to be merged or fast-forwarded",
		},
		{
			name:    "FailedWithSemanticAndRevertCommitsWhenDisabled",
			event:   "pullrequest:updated",
			prTitle: "fix: bananas",
			commits: append(semanticCommits(), map[string]interface{}{
				"message": "Revert \"feat: ride unicorns\"\n",
			}),
			config:              "commitsOnly: true\nallowRevertCommits: false",
			expectedState:       "FAILED",
			expectedDescription: "make sure every commit is semantic",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, &tc)
		})
	}
}

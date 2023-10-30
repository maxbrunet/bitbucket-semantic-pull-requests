package handler_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/maxbrunet/bitbucket-semantic-pull-requests/internal/handler"
)

func TestContains(t *testing.T) {
	t.Parallel()

	contains := handler.Contains([]string{"foo", "bar"}, "foo")

	require.True(t, contains)
}

func TestDoesNotContain(t *testing.T) {
	t.Parallel()

	contains := handler.Contains([]string{"foobar", "barfoo"}, "foo")

	require.False(t, contains)
}

package handler_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/maxbrunet/bitbucket-semantic-pull-requests/internal/handler"
)

func TestContains(t *testing.T) {
	contains := handler.Contains([]string{"foo", "bar"}, "foo")

	require.Equal(t, true, contains)
}

func TestDoesNotContain(t *testing.T) {
	contains := handler.Contains([]string{"foobar", "barfoo"}, "foo")

	require.Equal(t, false, contains)
}

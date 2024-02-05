package lib

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloneRepository(t *testing.T) {
	d := t.TempDir()
	err := GitCloneRepository(context.Background(), "https://github.com/sivukhin/govanish", "master", d)
	require.Nil(t, err)
}

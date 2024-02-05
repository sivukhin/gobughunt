package lib

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerExec(t *testing.T) {
	path, err := filepath.Abs("../")
	require.Nil(t, err)
	t.Log(path)
	lines, err := DockerExec(
		context.Background(),
		"sivukhinnikita/govanish:1.0.0",
		"/home",
		path,
	)
	require.Nil(t, err)
	t.Logf("%#v", lines)
}

func TestDockerStreamReader(t *testing.T) {
	t.Run("short", func(t *testing.T) {
		r := &DockerStreamReader{Reader: bytes.NewReader(
			append(
				append(
					[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05},
					[]byte("hello")...,
				),
				append(
					[]byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07},
					[]byte(", world")...,
				)...,
			),
		)}
		data, err := io.ReadAll(r)
		require.Nil(t, err)
		require.Equal(t, "hello, world", string(data))
	})
	t.Run("long", func(t *testing.T) {
		r := &DockerStreamReader{Reader: bytes.NewReader(
			append(
				[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01},
				[]byte(strings.Repeat("a", 257))...,
			),
		)}
		data, err := io.ReadAll(r)
		require.Nil(t, err)
		require.Equal(t, strings.Repeat("a", 257), string(data))
	})
	t.Run("broken", func(t *testing.T) {
		r := &DockerStreamReader{Reader: bytes.NewReader(
			append(
				[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01},
				[]byte(strings.Repeat("a", 128))...,
			),
		)}
		_, err := io.ReadAll(r)
		require.NotNil(t, err)
	})
}

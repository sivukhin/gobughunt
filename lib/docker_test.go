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
	t.Run("simple", func(t *testing.T) {
		path, err := filepath.Abs("../")
		require.Nil(t, err)
		lines, err := NaiveDockerApi.Exec(
			context.Background(),
			"sivukhinnikita/govanish:1.0.0@sha256:91fc7f5131aa71e5659de72b78934ecef3373cf1315469e5e8a9d3e18b7e0b89",
			"/home",
			path,
		)
		require.Nil(t, err)
		t.Logf("%#v", lines)
	})
	t.Run("non-zero exit code", func(t *testing.T) {
		path, err := filepath.Abs("../")
		require.Nil(t, err)
		_, err = NaiveDockerApi.Exec(
			context.Background(),
			"sivukhinnikita/dumb-fail:1.0.0@sha256:acc0726e21d1e9ea1c205216ad74c9d647b8f126d26af3586603462255fef969",
			"/home",
			path,
		)
		t.Log(err)
		require.ErrorIs(t, err, DockerNonZeroExitCodeErr)
	})
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

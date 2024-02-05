package lib

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func DockerExec(ctx context.Context, dockerImage string, containerBindPath, localBindPath string) ([]string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("unable to create docker client: %w", err)
	}
	pull, err := cli.ImagePull(ctx, dockerImage, types.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to pull docker image %v: %w", dockerImage, err)
	}
	for {
		n, err := io.Copy(io.Discard, pull)
		if n == 0 || err == io.EOF {
			break
		} else if err != nil {
			_ = pull.Close()
			return nil, fmt.Errorf("unable to pull docker image %v: %w", dockerImage, err)
		}
	}
	_ = pull.Close()

	containerConfig := &container.Config{Image: dockerImage, Cmd: []string{containerBindPath}}
	hostConfig := &container.HostConfig{Binds: []string{fmt.Sprintf("%v:%v", localBindPath, containerBindPath)}}
	create, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("unable to create container for image %v: %w", dockerImage, err)
	}
	attach, err := cli.ContainerAttach(ctx, create.ID, types.ContainerAttachOptions{
		Stream: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to attach to container %v: %w", create.ID, err)
	}
	defer attach.Close()

	err = cli.ContainerStart(ctx, create.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to start container %v: %w", create.ID, err)
	}
	lines := make([]string, 0)
	scanner := bufio.NewScanner(&DockerStreamReader{Reader: attach.Reader})
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to read stdout of container %v: %w", create.ID, err)
	}
	return lines, nil
}

type DockerStreamReader struct {
	Reader io.Reader
	chunk  []byte
}

func (r *DockerStreamReader) Read(p []byte) (int, error) {
	if len(r.chunk) == 0 {
		var buffer [8]byte
		_, err := io.ReadFull(r.Reader, buffer[:])
		if err != nil {
			return 0, err
		}
		size := binary.BigEndian.Uint32(buffer[4:])
		r.chunk = make([]byte, size)
		_, err = io.ReadFull(r.Reader, r.chunk)
		if err != nil {
			return 0, err
		}
	}
	n := len(p)
	if len(r.chunk) < n {
		n = len(r.chunk)
	}
	copy(p[:n], r.chunk[:n])
	r.chunk = r.chunk[n:]
	return n, nil
}

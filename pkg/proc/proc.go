package proc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"aquareum.tv/aquareum/pkg/log"
	"golang.org/x/sync/errgroup"
)

func RunMistServer(ctx context.Context) error {
	myself, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(myself, "MistServer")
	cmd.Env = []string{
		"MIST_NO_PRETTY_LOGGING=true",
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	group, _ := errgroup.WithContext(ctx)
	output := fmt.Println
	for i, pipe := range []io.ReadCloser{stdout, stderr} {
		func(i int, pipe io.ReadCloser) {
			group.Go(func() error {
				reader := bufio.NewReader(pipe)

				for {
					line, isPrefix, err := reader.ReadLine()
					if err != nil {
						if !errors.Is(err, io.EOF) {
							output(fmt.Sprintf("reader gave error, ending logging for fd=%d err=%s", i+1, err))
						}
						line, _, err := reader.ReadLine()
						if string(line) != "" {
							output(string(line))
						}
						return err
					}
					if isPrefix {
						output("warning: preceding line exceeds 64k logging limit and was split")
					}
					if string(line) != "" {
						level, procName, pid, path, streamName, msg, err := ParseMistLog(string(line))
						if err != nil {
							log.Log(ctx, "badly formatted mist log", "message", string(line))
						} else {
							log.Log(ctx, msg,
								"level", level,
								"procName", procName,
								"pid", pid,
								"streamName", streamName,
								"caller", path,
							)
						}
					}
				}
			})
		}(i, pipe)
	}

	group.Go(func() error {
		return cmd.Start()
	})

	return group.Wait()
}

func ParseMistLog(str string) (string, string, string, string, string, string, error) {
	parts := strings.Split(str, "|")
	if len(parts) != 6 {
		return "", "", "", "", "", "", fmt.Errorf("badly formatted mist string")
	}
	return parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], nil
}

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Stream(user string) error {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:9090/playback/%s/stream.mkv", user))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http status %s", resp.Status)
	}
	io.Copy(os.Stdout, resp.Body)
	return nil
}

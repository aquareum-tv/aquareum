package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Stream(u string) error {
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http status %s", resp.Status)
	}
	io.Copy(os.Stdout, resp.Body)
	return nil
}

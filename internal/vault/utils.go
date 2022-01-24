package vault

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

func RunWithStderr(cmd *exec.Cmd) error {
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errors.New(strings.Trim(errb.String(), "\n"))
		}
		return err
	}
	return nil
}

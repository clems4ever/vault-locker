package vault

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Sealer struct {
	keyFilePath string
	name        string
	devicePath  string
}

func NewSealer(name, devicePath, keyFilePath string) Sealer {
	return Sealer{
		keyFilePath: keyFilePath,
		name:        name,
		devicePath:  devicePath,
	}
}

func (v Sealer) Seal() error {
	return RunWithStderr(exec.Command("cryptsetup", "close", v.name))
}

func (v Sealer) Unseal() error {
	if _, err := os.Stat(v.devicePath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("device %s does not exist", v.devicePath)
	}

	cmd := exec.Command("cryptsetup", "open", v.devicePath, v.name, "--key-file="+v.keyFilePath)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdin, "")
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errors.New(strings.Trim(errb.String(), "\n"))
		}
		return err
	}
	return nil
}

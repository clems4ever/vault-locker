package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/godbus/dbus/v5"
)

// ErrDialogCanceled error representing the user cancelling the dialog
var ErrDialogCanceled = errors.New("User canceled dialog")

// ErrProcessExitedAbnormally represent an abnormal termination of the process
var ErrProcessExitedAbnormally = errors.New("Process exited abnormally")

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		panic(err)
	}

	conn.AddMatchSignal(
		dbus.WithMatchInterface("com.clems4ever.Vault"),
		dbus.WithMatchMember("OnMounted"))

	conn.AddMatchSignal(
		dbus.WithMatchInterface("com.clems4ever.Vault"),
		dbus.WithMatchMember("OnUnmounted"))

	signals := make(chan *dbus.Signal)
	conn.Signal(signals)

	fmt.Println("Waiting for events...")
	for sig := range signals {
		if sig.Name == "com.clems4ever.Vault.OnMounted" {
			onMounted()
		} else if sig.Name == "com.clems4ever.Vault.OnUnmounted" {
			onUnmounted()
		}
	}
}

func unsealVault(secret string) error {
	cmd := exec.Command("cryptsetup", "open", "/dev/vault", "vault")
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
	_, err = io.WriteString(stdin, secret)
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

func runWithStderr(cmd *exec.Cmd) error {
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

func sealVault() error {
	return runWithStderr(exec.Command("cryptsetup", "close", "vault"))
}

func mountVault() error {
	return runWithStderr(exec.Command("mount", "/dev/mapper/vault", "/vault"))
}

func unmountVault() error {
	return runWithStderr(exec.Command("umount", "/vault"))
}

func showPasswordDialog() (string, error) {
	secretBytes, err := exec.Command("qarma", "--password").Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return "", ErrDialogCanceled
			}
		}
		return "", err
	}
	return strings.Trim(string(secretBytes), "\n"), nil
}

func mountVaultProcess() error {
	var err error
	var secret string
	secret, err = showPasswordDialog()
	if err != nil {
		return err
	}

	fmt.Println("Unsealing vault...")
	err = unsealVault(secret + "\n")
	if err != nil {
		dialogErr := showErrorDialog(fmt.Sprintf("Unable to unseal vault: %s", err))
		if dialogErr != nil {
			fmt.Printf("Unable to show error dialog: %v", dialogErr)
		}
		return fmt.Errorf("Unable to unseal vault: %v", err)
	}

	fmt.Println("Mounting vault...")
	if err := mountVault(); err != nil {
		dialogErr := showErrorDialog(fmt.Sprintf("Unable to mount vault: %s", err))
		if dialogErr != nil {
			fmt.Printf("Unable to show error dialog: %v", dialogErr)
		}
		return fmt.Errorf("Unable to mount vault: %v", err)
	}

	err = showInfoDialog("Vault is mounted!")
	if err != nil {
		return err
	}
	return nil
}

func onMounted() {
	err := mountVaultProcess()
	if err != nil {
		fmt.Println(err)
	}

	for err != ErrDialogCanceled && err != nil {
		err = mountVaultProcess()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func showErrorDialog(message string) error {
	return exec.Command("qarma", "--error", "--text", message).Run()
}

func showInfoDialog(message string) error {
	return exec.Command("qarma", "--info", "--text", message).Run()
}

func onUnmounted() {

	var unmountingErr, sealingErr error
	if err := unmountVault(); err != nil {
		// If device is already unmounted we don't show error
		if err.Error() != "umount: /vault: not mounted." {
			unmountingErr = err
			fmt.Printf("Unable to unmount vault: %s\n", err)
		}
	}

	if err := sealVault(); err != nil {
		if err.Error() != "Device vault is not active." {
			sealingErr = err
			fmt.Printf("Unable to seal vault: %s\n", err)
		}
	}

	if unmountingErr != nil && sealingErr != nil {
		dialogErr := showErrorDialog(fmt.Sprintf("Unable to unmount and seal vault: %v, %v", unmountingErr, sealingErr))
		if dialogErr != nil {
			fmt.Printf("Unable to show error dialog: %v", dialogErr)
		}
	} else if unmountingErr != nil {
		dialogErr := showErrorDialog(fmt.Sprintf("Unable to unmount vault: %v", unmountingErr))
		if dialogErr != nil {
			fmt.Printf("Unable to show error dialog: %v", dialogErr)
		}
	} else if sealingErr != nil {
		dialogErr := showErrorDialog(fmt.Sprintf("Unable to unmount vault: %v", sealingErr))
		if dialogErr != nil {
			fmt.Printf("Unable to show error dialog: %v", dialogErr)
		}
	}
}

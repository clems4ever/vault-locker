package vault

import "os/exec"

func ShowErrorDialog(message string) error {
	return exec.Command("qarma", "--error", "--text", message).Run()
}

func ShowInfoDialog(message string) error {
	return exec.Command("qarma", "--info", "--text", message).Run()
}

/*func showPasswordDialog() (string, error) {
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
}*/

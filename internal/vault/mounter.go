package vault

import (
	"fmt"
	"os/exec"
)

type Mounter struct{}

func NewMounter() Mounter {
	return Mounter{}
}

func (m *Mounter) Mount(device, mountpoint string) error {
	if err := RunWithStderr(exec.Command("mount", device, mountpoint)); err != nil {
		return fmt.Errorf("unable to mount %s on %s: %v", device, mountpoint, err)
	}
	return nil
}

func (m *Mounter) Unmount(mountpoint string) error {
	err := RunWithStderr(exec.Command("umount", mountpoint))

	if err != nil {
		if err.Error() == "umount: /vault: not mounted." {
			return nil
		}
		return fmt.Errorf("unable to unmount vault: %s", err)
	}
	return nil
}

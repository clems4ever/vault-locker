package vault

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

type Vault struct {
	mounter          Mounter
	sealer           Sealer
	deviceMapperPath string
	mountpointPath   string
}

func NewVault(keyFilePath string) Vault {
	v := Vault{
		mounter:          NewMounter(),
		sealer:           NewSealer("vault", "/dev/vault", keyFilePath),
		deviceMapperPath: "/dev/mapper/vault",
		mountpointPath:   "/vault",
	}

	return v
}

func (v Vault) Seal() error {
	if err := v.mounter.Unmount(v.mountpointPath); err != nil {
		return fmt.Errorf("unable to unmount %s: %w", v.mountpointPath, err)
	}

	if err := v.sealer.Seal(); err != nil {
		if err.Error() != "Device vault is not active." {
			return fmt.Errorf("unable to seal vault: %w", err)
		}
	}
	return nil
}

func (v Vault) Unseal() error {
	err := v.sealer.Unseal( /* secret + "\n" */ )
	if err != nil {
		return fmt.Errorf("unable to unseal vault: %v", err)
	}

	err = v.mounter.Mount(v.deviceMapperPath, v.mountpointPath)
	if err != nil {
		return fmt.Errorf("unable to mount %s to %s: %v", v.deviceMapperPath, v.mountpointPath, err)
	}
	return nil
}

func mountedByStat(path string) (bool, error) {
	var st unix.Stat_t

	if err := unix.Lstat(path, &st); err != nil {
		return false, &os.PathError{Op: "stat", Path: path, Err: err}
	}
	dev := st.Dev
	parent := filepath.Dir(path)
	if err := unix.Lstat(parent, &st); err != nil {
		return false, &os.PathError{Op: "stat", Path: parent, Err: err}
	}
	if dev != st.Dev {
		// Device differs from that of parent,
		// so definitely a mount point.
		return true, nil
	}
	// NB: this does not detect bind mounts on Linux.
	return false, nil
}

func (v Vault) IsUnsealed() (bool, error) {
	return mountedByStat(v.mountpointPath)
}

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clems4ever/vault-unlocker/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var KeyFilePath = "/root/vault.bin"

// ErrDialogCanceled error representing the user cancelling the dialog
var ErrDialogCanceled = errors.New("user canceled dialog")

// ErrProcessExitedAbnormally represent an abnormal termination of the process
var ErrProcessExitedAbnormally = errors.New("process exited abnormally")

var rootCmd = &cobra.Command{
	Use:   "vault-locker",
	Short: "vault-locker lock and unlock a secure vault",
}

var unsealCmd = &cobra.Command{
	Use:   "unseal",
	Short: "unseal the vault",
	Run: func(cmd *cobra.Command, args []string) {
		unseal(vault.NewVault(KeyFilePath))
	},
}

var sealCmd = &cobra.Command{
	Use:   "seal",
	Short: "seal the vault",
	Run: func(cmd *cobra.Command, args []string) {
		seal(vault.NewVault(KeyFilePath))
	},
}

var isUnsealedCmd = &cobra.Command{
	Use:   "unsealed",
	Short: "check if vault is sealed",
	Run: func(cmd *cobra.Command, args []string) {
		ok, err := vault.NewVault(KeyFilePath).IsUnsealed()
		if err != nil {
			logrus.Panic(err)
		}
		if ok {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "run vault-locker in daemon mode",
	Run: func(cmd *cobra.Command, args []string) {
		d, err := cmd.PersistentFlags().GetDuration("autoseal-duration")
		if err != nil {
			logrus.Errorf("unable to parse autoseal-duration: %s", err)
		}
		RunDaemon(d)
	},
}

func main() {
	daemonCmd.PersistentFlags().Duration("autoseal-duration", 60*time.Second, "time before the vault is automatically sealed")
	rootCmd.AddCommand(unsealCmd, sealCmd, isUnsealedCmd, daemonCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type MountListener struct {
	vault vault.Vault
}

func NewMountListener(v vault.Vault) MountListener {
	return MountListener{
		vault: vault.NewVault(KeyFilePath),
	}
}
func (ml MountListener) OnUnseal() {
	unseal(ml.vault)
}
func (ml MountListener) OnSeal() {
	seal(ml.vault)
}

func RunDaemon(autoSealDuration time.Duration) {
	bus, err := vault.NewBus()
	if err != nil {
		logrus.Panic(err)
	}

	v := vault.NewVault(KeyFilePath)

	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	// Run the autoseal routine in background
	go vault.AutoSealRoutine(v, autoSealDuration)

	listener := NewMountListener(v)
	bus.Subscribe(listener)

	logrus.Infof("daemon is running with an autoseal duration of %s", autoSealDuration.String())
	// run the bus listener
	go bus.Listen()

	sig := <-cancelChan
	log.Printf("Caught SIGTERM %v", sig)

	err = v.Seal()
	if err != nil {
		logrus.Errorf("unable to unseal vault: %s", err)
	}
	logrus.Info("vault has been sealed automatically before quitting")
}

func unseal(v vault.Vault) {
	err := v.Unseal()
	if err != nil {
		dialogErr := vault.ShowErrorDialog(err.Error())
		if dialogErr != nil {
			logrus.Error("unable to show error dialog: %s", dialogErr)
		}
	} else {
		logrus.Info("vault has been unsealed")
		vault.ShowInfoDialog("vault is unsealed!")
	}
}

func seal(v vault.Vault) {
	err := v.Seal()
	if err != nil {
		dialogErr := vault.ShowErrorDialog(err.Error())
		if dialogErr != nil {
			logrus.Error("unable to show error dialog: %v", dialogErr)
		}
	} else {
		logrus.Info("vault has been sealed")
		vault.ShowInfoDialog("vault is sealed!")
	}
}

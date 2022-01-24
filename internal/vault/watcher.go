package vault

import (
	"time"

	"github.com/sirupsen/logrus"
)

func AutoSealRoutine(v Vault, autoSealDuration time.Duration) {
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(autoSealDuration)
	timer.Stop()

	autoUnsealTimerEnabled := false
	for {
		select {
		case <-ticker.C:
			unsealed, err := v.IsUnsealed()
			if err != nil {
				logrus.Errorf("unable to verify if vault is unsealed: %s", err)
				continue
			}
			if unsealed && !autoUnsealTimerEnabled {
				logrus.Infof("unsealed vault detected! autoseal will happen in %s", autoSealDuration)
				timer.Reset(autoSealDuration)
				autoUnsealTimerEnabled = true
			} else if !unsealed && autoUnsealTimerEnabled {
				logrus.Info("vault has been sealed manually!")
				timer.Stop()
				autoUnsealTimerEnabled = false
			}
		case <-timer.C:
			unsealed, err := v.IsUnsealed()
			if err != nil {
				logrus.Errorf("unable to verify if vault is unsealed: %s", err)
				continue
			}
			if unsealed {
				err = v.Seal()
				if err != nil {
					logrus.Errorf("unable to auto-seal vault: %s", err)
					continue
				}
				logrus.Info("vault has been auto-sealed!")
			}
			autoUnsealTimerEnabled = false
		}
	}
}

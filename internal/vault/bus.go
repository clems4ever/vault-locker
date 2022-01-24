package vault

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

type Bus struct {
	connection *dbus.Conn
	listeners  []MountListener
}

type MountListener interface {
	OnSeal()
	OnUnseal()
}

func NewBus() (*Bus, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("unable to listen to dbus: %w", err)
	}
	conn.AddMatchSignal(
		dbus.WithMatchInterface("com.clems4ever.Vault"),
		dbus.WithMatchMember("Unseal"))

	conn.AddMatchSignal(
		dbus.WithMatchInterface("com.clems4ever.Vault"),
		dbus.WithMatchMember("Seal"))
	return &Bus{connection: conn}, nil
}

func (b *Bus) Subscribe(listener MountListener) {
	b.listeners = append(b.listeners, listener)
}

func (b *Bus) Listen() {
	signals := make(chan *dbus.Signal)
	b.connection.Signal(signals)
	fmt.Println("Waiting for events...")
	for sig := range signals {
		if sig.Name == "com.clems4ever.Vault.Unseal" {
			for _, listener := range b.listeners {
				listener.OnUnseal()
			}
		} else if sig.Name == "com.clems4ever.Vault.Seal" {
			for _, listener := range b.listeners {
				listener.OnSeal()
			}
		}
	}
}

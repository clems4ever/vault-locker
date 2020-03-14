# Vault Unlocker

Application unsealing a LUKS vault when it is plugged to the computer
and re-sealing it when it's unplugged.

It leverages D-Bus to detect when the vault is plugged and unplugged,
open a Qt based input dialog box to enter the passphrase, unseal the
vault with cryptsetup and mount it at /vault.

## Dependencies

* Go
* cryptsetup command
* mount command
* qarma (https://github.com/luebking/qarma)

## Tests

Tested on Archlinux

# peugeot-tools

**Work in progress**

Go Command line to identify and download new versions of RCC/NAC Firmware and
MAP files in order to upgrade your Peugeot Car In-vehicle Infotainment.

# Motivation

`Peugeot Update` softwares provide only Microsoft and Apple updater.
Unfortunately, it does not offer alternative for the Free Software users (Linux,
FreeBSD, etc.) to assist them in downloading updates for their System Navigation
(NAC) and Firmware (RCC) using their favorite operating systems.

After some reverse engineering of the upgrade process in these applications, I
was able to reproduce the operations to retrieve new versions using Go and Free
Software! 

After downloading the new files, you just need to prepare an USB drive as
mentioned in [fr][fr] or [en][en] PDF instructions by Peugeot, then plug it in
your vehicule and launch the upgrade.

Enjoy upgrading your Peugeot Car!

# Build

As every simple golang project, it's just this command:

```
go build
```

# Usage

```shell
peugeot-tools -vin "VFX..."
```

# License

BSD License

[fr]: https://media.peugeot.fr/file/21/5/fr-notice-d-utilisation-rcc-softv3.476215.pdf
[en]: https://media.peugeot.co.uk/file/78/0/instructions-for-updating-the-rcc-software-system.489780.pdf

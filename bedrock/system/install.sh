#!/bin/sh

set -e

install_service() {
	unit=$1
	ln -fs /root/system/$unit.timer /etc/systemd/system
	ln -fs /root/system/$unit.service /etc/systemd/system
}

enable_service() {
	unit=$1
	systemctl enable $unit.timer
	systemctl start $unit.timer
}

install_service start-mc-server
install_service stop-mc-server
install_service sync-member

systemctl daemon-reload

enable_service start-mc-server
enable_service stop-mc-server
enable_service sync-member

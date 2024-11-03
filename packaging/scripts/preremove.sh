#!/bin/sh

# Step 1, decide if we should use SystemD or init/upstart
use_systemctl="True"
systemd_version=0
if ! command -V systemctl >/dev/null 2>&1; then
  use_systemctl="False"
fi


if [ "${use_systemctl}" = "True" ]; then
	printf "\033[32m Stop the service and timer units\033[0m\n"
	systemctl stop readflow.timer ||:
	systemctl stop readflow.service ||:
	printf "\033[32m Set the disabled flag for the service and timer units\033[0m\n"
	systemctl disable readflow.service ||:
	systemctl disable readflow.timer ||:
	printf "\033[32m Mask the service and timer\033[0m\n"
	systemctl mask readflow.service ||:
	systemctl mask readflow.timer ||:
	printf "\033[32m Reload the service unit from disk\033[0m\n"
	systemctl daemon-reload ||:
fi

#!/usr/bin/env sh

set -e

cronjob="$CRON_SCHEDULE /readflow sync"
echo "$cronjob" > /crontab
chmod 644 /crontab

# first arg is `-f` or `--some-option`
if [ "${1#-}" != "$1" ] || [ "$1" == "sync" ]; then
	set -- /readflow "$@"
fi

exec "$@"


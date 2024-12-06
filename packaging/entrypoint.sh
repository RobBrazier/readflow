#!/usr/bin/env sh

set -e

cronjob="$CRON_SCHEDULE /bin/readflow sync"
echo "$cronjob" > /tmp/crontab
chmod 644 /tmp/crontab

# first arg is `-f` or `--some-option`
if [ "${1#-}" != "$1" ] || [ "$1" == "sync" ]; then
	set -- /bin/readflow "$@"
fi

exec "$@"


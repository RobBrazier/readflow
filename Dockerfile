FROM docker.io/alpine:20240923
ENV READFLOW_DOCKER="1" \
	SOURCE="database" \
	COLUMN_CHAPTER="false" \
	TARGETS="anilist,hardcover" \
	DATABASE_CALIBRE="/data/metadata.db" \
	DATABASE_CALIBREWEB="/data/app.db" \
	CRON_SCHEDULE="@hourly"

COPY packaging/entrypoint.sh /
RUN chmod +x /entrypoint.sh && \
	apk add --no-cache supercronic
COPY readflow /bin

ENTRYPOINT ["/entrypoint.sh"]

CMD ["supercronic", "-no-reap", "/tmp/crontab"]

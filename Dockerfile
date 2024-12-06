FROM docker.io/alpine:3
ENV READFLOW_DOCKER="1"
ENV SOURCE="database"
ENV COLUMN_CHAPTER="false"
ENV TARGETS="anilist,hardcover"

ENV DATABASE_CALIBRE="/data/metadata.db"
ENV DATABASE_CALIBREWEB="/data/app.db"
ENV CRON_SCHEDULE="@hourly"

COPY packaging/entrypoint.sh /
RUN chmod +x /entrypoint.sh && \
	apk add --no-cache supercronic
COPY readflow /bin

ENTRYPOINT ["/entrypoint.sh"]

CMD ["supercronic", "-no-reap", "/tmp/crontab"]

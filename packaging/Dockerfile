FROM scratch

ENV READFLOW_DOCKER="1" \
	SOURCE="database" \
	COLUMN_CHAPTER="false" \
	TARGETS="anilist,hardcover" \
	DATABASE_CALIBRE="/data/metadata.db" \
	DATABASE_CALIBREWEB="/data/app.db" \
	CRON_SCHEDULE="@hourly"

COPY readflow /

ENTRYPOINT ["/readflow"]

CMD ["schedule"]

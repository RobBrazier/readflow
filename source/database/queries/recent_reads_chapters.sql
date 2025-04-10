SELECT
    b.id as book_id,
    b.title AS book_name,
    s.id AS series_id,
    b.series_index as book_series_index,
    i_anilist.val as anilist_id,
    i_hardcover.val as hardcover_id,
	i_hardcover_edition.val AS hardcover_edition_id,
    kb.progress_percent as progress_percent,
    c.value as chapter_count
FROM
    book_read_link brl
LEFT JOIN
    calibre.books b ON b.id = brl.book_id
LEFT JOIN
    calibre.books_series_link bsl ON bsl.book = b.id
LEFT JOIN
    calibre.series s ON bsl.series = s.id
LEFT JOIN
    calibre.identifiers i_anilist ON i_anilist.book = b.id AND i_anilist.type = 'anilist'
LEFT JOIN
    calibre.identifiers i_hardcover ON i_hardcover.book = b.id AND i_hardcover.type = 'hardcover'
LEFT JOIN
	calibre.identifiers i_hardcover_edition ON i_hardcover_edition.book = b.id AND i_hardcover_edition.type = 'hardcover-edition'
LEFT JOIN
    calibre.%s c ON c.book = b.id
LEFT JOIN
    kobo_reading_state krs ON krs.book_id = b.id
LEFT JOIN
    kobo_bookmark kb ON kb.kobo_reading_state_id = krs.id
WHERE
    krs.last_modified > datetime('now', $1)
    AND kb.progress_percent IS NOT NULL
    AND brl.read_status != 0;


package source

const CHAPTERS_QUERY = "SELECT id FROM calibre.custom_columns WHERE label = ?;"

// I'd love to compress these into one query, but can't figure out a SQL query that still passes if a table doesn't exist
const RECENT_READS_QUERY = `
SELECT
    b.id as book_id,
    b.title AS book_name,
    s.id AS series_id,
    b.series_index as book_series_index,
    i_isbn.val as isbn,
    i_anilist.val as anilist_id,
    i_hardcover.val as hardcover_id,
    i_hardcover_edition.val as hardcover_edition,
    kb.progress_percent as progress_percent,
    brl.read_status as read_status,
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
    calibre.identifiers i_isbn ON i_isbn.book = b.id AND i_isbn.type = 'isbn'
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
    krs.last_modified > datetime('now', ?)
    AND kb.progress_percent IS NOT NULL
    AND brl.read_status != 0;
`
const RECENT_READS_QUERY_RANKED = `
WITH ranked_books AS (
    SELECT
        b.id as book_id,
        b.title AS book_name,
        s.id AS series_id,
        b.series_index as book_series_index,
		i_isbn.val as isbn,
        i_anilist.val as anilist_id,
        i_hardcover.val as hardcover_id,
        i_hardcover_edition.val as hardcover_edition,
        kb.progress_percent as progress_percent,
        brl.read_status as read_status,
        c.value as chapter_count,
        CASE 
            WHEN s.id IS NOT NULL THEN ROW_NUMBER() OVER (PARTITION BY s.id ORDER BY b.series_index DESC)
            ELSE 1
        END as rn
    FROM
        book_read_link brl
    LEFT JOIN
        calibre.books b ON b.id = brl.book_id
    LEFT JOIN
        calibre.books_series_link bsl ON bsl.book = b.id
    LEFT JOIN
        calibre.series s ON bsl.series = s.id
	LEFT JOIN
        calibre.identifiers i_isbn ON i_isbn.book = b.id AND i_isbn.type = 'isbn'
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
        krs.last_modified > datetime('now', ?)
        AND kb.progress_percent IS NOT NULL
        AND brl.read_status != 0
) SELECT
    book_id,
    book_name,
    series_id,
    book_series_index,
    read_status,
	isbn,
    anilist_id,
    hardcover_id,
    hardcover_edition,
    progress_percent,
    chapter_count
FROM
    ranked_books
WHERE
    rn = 1
    AND (anilist_id IS NOT NULL OR hardcover_id IS NOT NULL)
ORDER BY
    series_id NULLS LAST, book_series_index DESC;
`

const RECENT_READS_QUERY_NO_CHAPTERS = `
SELECT
    b.id as book_id,
    b.title AS book_name,
    s.id AS series_id,
    b.series_index as book_series_index,
    i_isbn.val as isbn,
    i_anilist.val as anilist_id,
    i_hardcover.val as hardcover_id,
    i_hardcover_edition.val as hardcover_edition,
    kb.progress_percent as progress_percent,
    brl.read_status as read_status,
    NULL as chapter_count
FROM
    book_read_link brl
LEFT JOIN
    calibre.books b ON b.id = brl.book_id
LEFT JOIN
    calibre.books_series_link bsl ON bsl.book = b.id
LEFT JOIN
    calibre.series s ON bsl.series = s.id
LEFT JOIN
    calibre.identifiers i_isbn ON i_isbn.book = b.id AND i_isbn.type = 'isbn'
LEFT JOIN
    calibre.identifiers i_anilist ON i_anilist.book = b.id AND i_anilist.type = 'anilist'
LEFT JOIN
    calibre.identifiers i_hardcover ON i_hardcover.book = b.id AND i_hardcover.type = 'hardcover'
LEFT JOIN
    calibre.identifiers i_hardcover_edition ON i_hardcover_edition.book = b.id AND i_hardcover_edition.type = 'hardcover-edition'
LEFT JOIN
    kobo_reading_state krs ON krs.book_id = b.id
LEFT JOIN
    kobo_bookmark kb ON kb.kobo_reading_state_id = krs.id
WHERE
    krs.last_modified > datetime('now', ?)
    AND kb.progress_percent IS NOT NULL
    AND brl.read_status != 0
`
const RECENT_READS_QUERY_NO_CHAPTERS_RANKED = `
WITH ranked_books AS (
    SELECT
        b.id as book_id,
        b.title AS book_name,
        s.id AS series_id,
        b.series_index as book_series_index,
		i_isbn.val as isbn,
        i_anilist.val as anilist_id,
        i_hardcover.val as hardcover_id,
        i_hardcover_edition.val as hardcover_edition,
        kb.progress_percent as progress_percent,
        brl.read_status as read_status,
        NULL as chapter_count,
        CASE 
            WHEN s.id IS NOT NULL THEN ROW_NUMBER() OVER (PARTITION BY s.id ORDER BY b.series_index DESC)
            ELSE 1
        END as rn
    FROM
        book_read_link brl
    LEFT JOIN
        calibre.books b ON b.id = brl.book_id
    LEFT JOIN
        calibre.books_series_link bsl ON bsl.book = b.id
    LEFT JOIN
        calibre.series s ON bsl.series = s.id
	LEFT JOIN
        calibre.identifiers i_isbn ON i_isbn.book = b.id AND i_isbn.type = 'isbn'
    LEFT JOIN
        calibre.identifiers i_anilist ON i_anilist.book = b.id AND i_anilist.type = 'anilist'
    LEFT JOIN
        calibre.identifiers i_hardcover ON i_hardcover.book = b.id AND i_hardcover.type = 'hardcover'
    LEFT JOIN
        calibre.identifiers i_hardcover_edition ON i_hardcover_edition.book = b.id AND i_hardcover_edition.type = 'hardcover-edition'
    LEFT JOIN
        kobo_reading_state krs ON krs.book_id = b.id
    LEFT JOIN
        kobo_bookmark kb ON kb.kobo_reading_state_id = krs.id
    WHERE
        krs.last_modified > datetime('now', ?)
        AND kb.progress_percent IS NOT NULL
        AND brl.read_status != 0
) SELECT
    book_id,
    book_name,
    series_id,
    book_series_index,
    read_status,
	isbn,
    anilist_id,
    hardcover_id,
    hardcover_edition,
    progress_percent,
    chapter_count
FROM
    ranked_books
WHERE
    rn = 1
    AND (anilist_id IS NOT NULL OR hardcover_id IS NOT NULL)
ORDER BY
    series_id NULLS LAST, book_series_index DESC;
`

const PREVIOUS_BOOKS_IN_SERIES_QUERY = `
SELECT 
    b.id AS book_id,
    b.title AS book_name,
	bsl.series AS series_id,
    b.series_index AS book_series_index,
	brl.read_status as read_status,
	i_isbn.val AS isbn,
    i_anilist.val AS anilist_id,
	i_hardcover.val AS hardcover_id,
    i_hardcover_edition.val as hardcover_edition,
	NULL AS progress_percent,
    c.value AS chapter_count
FROM 
    calibre.books b
INNER JOIN 
    calibre.books_series_link bsl ON b.id = bsl.book
LEFT JOIN
	calibre.identifiers i_isbn ON i_isbn.book = b.id AND i_isbn.type = 'isbn'
LEFT JOIN
	calibre.identifiers i_anilist ON i_anilist.book = b.id AND i_anilist.type = 'anilist'
LEFT JOIN
	calibre.identifiers i_hardcover ON i_hardcover.book = b.id AND i_hardcover.type = 'hardcover'
LEFT JOIN
    calibre.identifiers i_hardcover_edition ON i_hardcover_edition.book = b.id AND i_hardcover_edition.type = 'hardcover-edition'
LEFT JOIN 
    calibre.%s c ON b.id = c.book
LEFT JOIN
	book_read_link brl ON brl.book_id = b.id
WHERE 
    bsl.series = ?
	AND b.series_index < ?
ORDER BY 
    b.series_index;
`

const PREVIOUS_BOOKS_IN_SERIES_QUERY_NO_CHAPTERS = `
SELECT 
    b.id AS book_id,
    b.title AS book_name,
	bsl.series AS series_id,
    b.series_index AS book_series_index,
	brl.read_status as read_status,
	i_isbn.val AS isbn,
    i_anilist.val AS anilist_id,
	i_hardcover.val AS hardcover_id,
    i_hardcover_edition.val as hardcover_edition,
	NULL AS progress_percent,
    NULL AS chapter_count
FROM 
    calibre.books b
INNER JOIN 
    calibre.books_series_link bsl ON b.id = bsl.book
LEFT JOIN
	calibre.identifiers i_isbn ON i_isbn.book = b.id AND i_isbn.type = 'isbn'
LEFT JOIN
	calibre.identifiers i_anilist ON i_anilist.book = b.id AND i_anilist.type = 'anilist'
LEFT JOIN
	calibre.identifiers i_hardcover ON i_hardcover.book = b.id AND i_hardcover.type = 'hardcover'
LEFT JOIN
    calibre.identifiers i_hardcover_edition ON i_hardcover_edition.book = b.id AND i_hardcover_edition.type = 'hardcover-edition'
LEFT JOIN
	book_read_link brl ON brl.book_id = b.id
WHERE 
    bsl.series = ?
	AND b.series_index < ?
ORDER BY 
    b.series_index;
`

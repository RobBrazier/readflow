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


SELECT 
    b.id AS book_id,
    b.title AS book_name,
	bsl.series AS series_id,
    b.series_index AS book_series_index,
    i_anilist.val AS anilist_id,
	i_hardcover.val AS hardcover_id,
	i_hardcover_edition.val AS hardcover_edition_id,
	NULL AS progress_percent,
    c.value AS chapter_count
FROM 
    calibre.books b
INNER JOIN 
    calibre.books_series_link bsl ON b.id = bsl.book
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
    bsl.series = $1
	AND b.series_index < $2
ORDER BY 
    b.series_index;


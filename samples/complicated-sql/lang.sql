-- +gen:sql v=0.3


-- 获取所有多语言的code和语言id
-- +gen:method GetLangs
SELECT
	code,
	languages_id Id
FROM
	languages
WHERE
	code IN ('en', 'cn', 'tw', 'th');

-- +gen:end



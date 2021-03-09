-- +gen:sql v=0.3


-- 列出所有的供应商放假时间(生效中)
-- +gen:method List
SELECT
	supplier_id Id,
	start_date,
	end_date
FROM
	supplier_festival_range
WHERE
	status=1 AND end_date>=${now}

-- +gen:end


-- 获取某个特定供应商的放假时间(生效中)
-- +gen:method Get
SELECT
	supplier_id Id,
	start_date,
	end_date
FROM
	supplier_festival_range
WHERE
	status=1 AND supplier_id=${id}
	AND end_date>=${now}

-- +gen:end

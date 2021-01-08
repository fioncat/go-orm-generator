-- !GetById
SELECT
	id, name, email, phone, age, create_time,
	password, is_admin
FROM user
WHERE is_delete=0 AND id=${id}


-- !List
SELECT
 id, name, email, phone, age, create_time,
 password, is_admin
FROM user
WHERE is_delete=0
LIMIT ${offset}, ${limit}

-- !Update
UPDATE user
SET name=${u.Name}, email=${u.Email},
    phone=${u.Phone}, age=${u.Age},
	password=${u.Password}
WHERE id=${u.Id}


-- !GetDetail
SELECT
  user_id, score, balance, text
FROM user_detail
WHERE user_id=${id}



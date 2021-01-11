-- UserDb的SQL语句

-- !FindById
SELECT
	id, name, email, phone, age, create_time,
	password, is_admin, is_delete

FROM user
WHERE id=${id} AND is_delete=0;


-- !FindAdmins
SELECT
	id, name, email, phone, age, create_time,
 	password, is_admin, is_delete

FROM user
WHERE is_admin=1 AND is_delete=0;


-- !UpdateAge
UPDATE user SET age=${age}
WHERE id=${id};


-- !DeleteUser
UPDATE user SET is_delete=1
WHERE id=${id};


-- !Count
SELECT COUNT(1) FROM user WHERE is_delete=0;


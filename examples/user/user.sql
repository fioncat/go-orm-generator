-- !add
INSERT INTO user(name, email, phone, age)
VALUES (${u.Name}, ${u.Email}, ${u.Phone}, ${u.Age})

-- !update
UPDATE user SET
  name=${u.Name},
  email=${u.Email},
  phone=${u.Phone},
  age=${u.Age}

WHERE id=${u.Id}

-- !findById
SELECT
	id,
	name,
	email,
	phone,
	age,
	create_time,
	password,
	is_admin
FROM user
WHERE id=${id}


-- !search
SELECT
	id,
	name,
	email,
	phone,
	age,
	create_time,
	password,
	is_admin
FROM user
WHERE is_delete=0 AND #{where}


-- !searchConds
SELECT
  id,
  name,
  email,
  phone,
  age,
  create_time,
  password,
  is_admin
FROM user
WHERE email=${email} OR phone=${phone}


-- !count
SELECT COUNT(1)
FROM user


-- !countAdmin
SELECT COUNT(1) FROM user
WHERE is_admin=${admin}


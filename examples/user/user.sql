-- !Add
INSERT INTO user(name, email, phone, age, password)
VALUES (${u.Name}, ${u.Email}, ${u.Phone}, ${u.Password});

-- !Update
UPDATE user SET name=${u.Name}, email=${u.Email}, phone=${u.Phone}, password=${u.Password}
WHERE id=${u.Id};

-- !FindByID
SELECT id, name, email, phone, age, password, create_time, is_admin
FROM user WHERE id=${id};

-- !FindByName
SELECT id, name, email, phone, age, password, create_time, is_admin
FROM user WHERE name=${name};

-- !Search
SELECT id, name, email, phone, age, password, create_time, is_admin
FROM user
WHERE 1=1 #{where}
LIMIT ${offset}, ${limit};

-- !Count
SELECT COUNT(1)
FROM user
WHERE 1=1 #{where};

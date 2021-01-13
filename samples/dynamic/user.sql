
-- +Update
UPDATE user SET 1=1
+{if u.Name != ""} AND name=${u.Name} +{endif}
+{if u.Age > 0} AND age=${u.Age} +{endif}
+{if u.Phone != ""} AND phone=${u.Phone} +{endif}
WHERE id=${u.Id}


-- +FindByIds
SELECT id, name
FROM user
WHERE id IN (+{for id:ids:','} ${id} +{endfor})


-- +BatchInsert
INSERT INTO user(id,name,age,phone,is_admin) VALUES
+{for u:users:','}
  (${u.Id}, ${u.Name}, ${u.Age}, ${u.Phone}, ${u.IsAdmin})
+{endfor}


-- !FindById
SELECT id, name
FROM user
WHERE id=${id}



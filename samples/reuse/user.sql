-- @userFields
id, name, age, is_admin

-- @searchUser
u.id, u.name, u.age, u.is_admin

-- @searchDetail
ud.text, ud.balance, ud.score

-- !FindById
SELECT @{userFields}
FROM user
WHERE id=${id}


-- +Detail
SELECT @{searchUser}, @{searchDetail}
FROM user u
JOIN user_detail ud ON u.id=ud.user_id
WHERE
+{for cond:conds:' AND'}
  #{cond.Key}=${cond.Val}
+{endfor}


-- +Search
SELECT @{searchUser}
FROM user u
WHERE
+{for cond:conds:' AND'}
  #{cond.Key}#{cond.Op}${cond.Val}
+{endfor}
LIMIT ${offset}, ${limit}


-- +Adds
INSERT INTO user(@{userFields})
VALUES
+{for u:users:,}
  (${u.Id}, ${u.Name}, ${u.Age}, ${u.IsAdmin})
+{endfor}



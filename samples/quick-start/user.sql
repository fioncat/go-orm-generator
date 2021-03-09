-- +gen:sql v=0.3

-- +gen:method UserOper.FindById dyn=false
SELECT id, name, age,
FROM user
WHERE id=${id}
-- +gen:end


-- +gen:method UserOper.Add
INSERT INTO user(id, name, age)
VALUES (${u.Id}, ${u.Name}, ${u.Age})
-- +gen:end




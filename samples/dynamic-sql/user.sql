-- +gen:sql v=0.3

-- +gen:method FindByCond dyn=true
SELECT
	id, name, email, phone,
	age, create_time, password,
	is_admin, is_delete
FROM
	user
WHERE 1=1
%{if conds["id"] != nil}
	AND id=${conds["id"]}
%{endif}
%{if conds["name"] != nil}
	AND name=${conds["name"]}
%{endif}
%{if conds["email"] != nil}
	AND email=${conds["email"]}
%{endif}
%{if conds["phone"] != nil}
	AND phone=${conds["phone"]}
%{endif}
LIMIT
	${offset}, ${limit}
-- +gen:end


-- +gen:method Adds dyn=true
INSERT INTO `user`(`name`, `email`, `phone`, `age`)
VALUES
%{for u in us join ','}
	  (${u.Name}, ${u.Email}, ${u.Phone}, ${u.Age})
%{endfor}

-- +gen:end


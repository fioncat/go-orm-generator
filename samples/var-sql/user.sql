-- +gen:sql v=0.3


-- +gen:var userFields
u.name, u.email, u.phone, u.text
-- +gen:end


-- +gen:var idCond
u.id=${id}
-- +gen:end


-- +gen:method FindById
SELECT
	@{userFields}
FROM
	user u
WHERE @{idCond}
-- +gen:end


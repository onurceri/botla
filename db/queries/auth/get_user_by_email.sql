-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name
FROM users
WHERE email = $1;


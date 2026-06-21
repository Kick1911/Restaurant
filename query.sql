-- name: CreateTenant :one
INSERT INTO tenants (id, name, slug, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: FindTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: CreateUser :one
INSERT INTO users (id, tenant_id, name, email, password_hash, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE tenant_id = $1 AND email = $2;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateDish :one
INSERT INTO dishes (id, tenant_id, name, description, price, image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING *;

-- name: FindDishByID :one
SELECT * FROM dishes WHERE tenant_id = $1 AND id = $2;

-- name: UpdateDish :one
UPDATE dishes SET name = $3, description = $4, price = $5, image_url = $6, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: DeleteDish :exec
DELETE FROM dishes WHERE tenant_id = $1 AND id = $2;

-- name: SearchDishes :many
SELECT * FROM dishes
WHERE tenant_id = $1
  AND (name ILIKE '%' || $2 || '%' OR $2 = '')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CreateRating :one
INSERT INTO ratings (id, dish_id, user_id, rating, created_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (dish_id, user_id) DO UPDATE SET rating = $4
RETURNING *;

-- name: FindRatingsByDishID :many
SELECT * FROM ratings WHERE dish_id = $1 ORDER BY created_at DESC;

-- name: GetAverageRating :one
SELECT COALESCE(AVG(rating), 0)::float FROM ratings WHERE dish_id = $1;

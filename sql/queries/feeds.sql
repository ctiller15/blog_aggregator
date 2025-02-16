-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetFeeds :many
SELECT feeds.name AS feed_name, feeds.url, users.name AS user_name
FROM feeds
INNER JOIN users
ON feeds.user_id = users.id;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (INSERT INTO feed_follows (
    id,
    created_at,
    updated_at,
    user_id,
    feed_id
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING  *
)

SELECT 
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
    FROM inserted_feed_follow
    INNER JOIN users
    ON users.id = inserted_feed_follow.user_id
    INNER JOIN feeds
    ON feeds.id = inserted_feed_follow.feed_id;

-- name: GetFeedByUrl :one
SELECT id, feeds.name AS feed_name
FROM feeds
WHERE url = $1
LIMIT 1;

-- name: GetFeedFollowsForUser :many
SELECT users.name as user_name, feeds.name as feed_name, users.id as user_id, feeds.id as feed_id
FROM users
INNER JOIN feed_follows
ON users.id = feed_follows.user_id
INNER JOIN feeds
ON feed_follows.feed_id = feeds.id
WHERE users.name = $1;

-- name: DeleteFeeds :exec
DELETE FROM feeds;
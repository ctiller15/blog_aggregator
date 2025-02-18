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

-- name: DeleteFeedFollowByUrl :exec
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1
AND feed_follows.feed_id IN (
    SELECT id
    FROM feeds
    WHERE feeds.url = $2
);

-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = current_timestamp,
updated_at = current_timestamp
WHERE feeds.id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST;

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.*
FROM posts
INNER JOIN feeds
ON posts.feed_id = feeds.id
INNER JOIN feed_follows
ON feeds.id = feed_follows.feed_id
INNER JOIN users
ON feed_follows.user_id = users.id
WHERE users.id = $1
ORDER BY posts.published_at DESC
LIMIT $2;
# Boot.dev blog aggregator

## Requirements
- postgres@15+
- golang@23.x+

### install Postgres
```bash
# macos
brew install postgresql@15

# linux/wsl
sudo apt update
sudo apt install postgresql postgresql-contrib
```

```bash
# in postgres
CREATE DATABASE gator;
```

### Install the gator CLI
```bash
# from home directory
go build .
```

### install goose (for migrations)
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Setup config
```bash
touch ~/.gatorconfig.json
```

```json
{
  "db_url": "postgres://example"
}
```


## Usage

### Commands
All commands can be run with `gator {command}`
"login" - logs in a user
"register" - registers a user
"reset" - resets the database
"users" - lists all users
"agg" - scrapes existing feeds at a given rate
"addfeed" - adds a feed
"feeds" - lists all feeds
"follow" - follows a feed as a user
"following" - lists feeds a user is following
"unfollow" - unfollows a feed for a user
"browse" - browses a given number of feeds

### Plumbing
```
# generate models
sqlc generate
```
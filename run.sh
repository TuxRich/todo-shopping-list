#!/bin/sh
set -e

export DB_PATH="/data/todo.db"
export PORT="8099"

exec /usr/local/bin/todo-app

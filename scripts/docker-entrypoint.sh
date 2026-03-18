#!/bin/sh

echo "Migrating the database..."
goose -dir ./sql/migrations postgres "$DATABASE_URL" up
echo "Migrations completed successfully!"

echo "Starting the server..."
exec /usr/local/bin/bookme
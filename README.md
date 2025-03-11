# The Virtual Armory

A web application for the gun community to track the weapons they own. Eventual improvements will include ability
to track ammo owned. Then range visits to track shots put down range per weapon. And more.

## Features

- User authentication (login, register, password recovery)
- Admin user support
- GORM for database abstraction
- Gin for routing
- Templ for templating

## Prerequisites

- Go 1.24+
- PostgreSQL
- Docker (optional, for running PostgreSQL)

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
PORT=8080
HOST=localhost
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_DATABASE=virtualarmory
DB_SCHEMA=public
```

## Running the Application

1. Start PostgreSQL:

```bash
docker-run
```

2. Run the application:

```bash
go run main.go
```

Or use the Makefile:

```bash
make run
```

## Creating an Admin User

To create an admin user, run:

```bash
go run cmd/scripts/create_admin.go -email <email> -password<password>
```

For example:

```bash
go run cmd/scripts/create_admin.go -emailadmin@example.com -password password123
```

## Authentication

The application uses Authboss for authentication. The following routes are available:

- `/login` - Login page
- `/register` - Registration page
- `/recover` - Password recovery page

## Protected Routes

The following routes are protected and require authentication:

- `/owner` - User armory page
- `/profile` - User profile page

## Admin Routes

The following routes are protected and require admin privileges:

- `/admin/dashboard` - Admin dashboard

## Development

To generate Templ files:

```bash
templ generate
```

To run tests:

```bash
make test
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```

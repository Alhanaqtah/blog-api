# Blog API

This is a simple blog API project implemented in Go, utilizing SQLite as the database and Chi as the router. The project follows a clean architecture pattern, separating concerns into handler, service, and data access layers.

## Features

- **Users:** CRUD operations for managing users, including registration, login, update, and removal.
- **Articles:** CRUD operations for managing articles, including creation, retrieval by ID, update, and removal.
- **Authentication:** Authentication system using JWT tokens.
- **Encryption:** Passwords are hashed using bcrypt for security.

## Configuration

The project uses a `config.yaml` file for configuration:

```yaml
env: "local"
storage_path: "./storage/storage.db"
secret: "secret"
http_server:
  address: "localhost:8080"
  timeout: 4s
  idle_timeout: 30s
  shutdown_timeout: 10s
  tokenTTL: 12h
```

## Setup

1. Clone the repository:

```
git clone https://github.com/Alhanaqtah/blog-api.git
```

2. Navigate to the project directory:

```
cd blog-api
```

3. Build and run the project:

```
make run
```

4. The API will be accessible at `http://localhost:8080`.

## Getting Started

To start using the API, you can use tools like Postman or cURL to make HTTP requests to the provided endpoints. Ensure to include proper authentication headers when accessing protected endpoints.

For example, to create a new article:

```
POST http://localhost:8080/articles

Request Body:
{
  "title": "New Article",
  "content": "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
}
```

For authentication, you can obtain a JWT token by logging in with valid credentials. This token should be included in the `Authorization` header of subsequent requests.

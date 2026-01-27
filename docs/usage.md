## Libraries & Dependencies 📦

Below is a rough list of the main libraries used in this project and **why they were chosen**.

This should give readers (and evaluators 👀) a quick mental model of the stack.

---

### Session & Security

```bash
go get github.com/gorilla/sessions
```

- Used for handling **HTTP sessions**
- Primarily required for **OAuth state management**
- Protects against **CSRF attacks** during OAuth redirects

```bash
go get github.com/gorilla/securecookie
```

- Used **indirectly** by `gorilla/sessions`
- Handles secure encoding and decoding of session cookies

---

### Authentication & Authorization

```bash
go get github.com/golang-jwt/jwt/v5
```

- Used for **JWT access tokens**
- Provides signing, validation, and claim parsing
- Keeps API endpoints stateless after OAuth login

```bash
go get golang.org/x/oauth2
```

- Official Go OAuth2 client
- Used for **42 Intra OAuth authentication flow**
- Handles token exchange and authenticated HTTP clients

---

### Database

```bash
go get github.com/lib/pq
```

- PostgreSQL driver for Go
- Required to be imported to **register the driver**
- No direct usage in code (used implicitly via `database/sql`)

```bash
go get github.com/google/uuid
```

- Used for generating **UUIDs**
- Helpful for identifiers, tokens, and internal references

---

### Configuration

```bash
go get github.com/joho/godotenv
```

- Loads environment variables from `.env` file
- Keeps secrets and config out of source control
- Useful for local development and testing

---

### Email

```bash
go get github.com/wneessen/go-mail
```

- SMTP client library
- Used for sending:
  - Booking confirmation emails
  - Cancellation notifications
- Supports TLS and modern SMTP features

---

## go.mod Overview 🧬

```go
module github.com/IbnBaqqi/book-me

go 1.25.5

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/sessions v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/wneessen/go-mail v0.7.2
	golang.org/x/oauth2 v0.34.0
)

require (
	github.com/gorilla/securecookie v1.1.2 // indirect
	golang.org/x/text v0.29.0 // indirect
)
```

---

### Why This Stack? 🤔

- **Pure `net/http`** – no heavy frameworks
- OAuth handled explicitly for learning and control
- PostgreSQL + SQL-first approach
- Clear separation between:
  - Authentication
  - Business rules
  - Infrastructure (email, DB, OAuth)

This keeps the project **simple, explicit, and easy to reason about**.

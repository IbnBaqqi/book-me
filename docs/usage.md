## Libraries & Dependencies

Below is a list of the main libraries used in this project and **why they were chosen**.

---

### Validation

```bash
go get github.com/go-playground/validator/v10
```

- Used for **struct validation** with custom rules
- Validates reservation times, date ranges, and business rules
- Custom validators for:
  - Future time validation
  - School hours (6 AM - 8 PM)
  - Maximum date range (60 days)

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

### Google Calendar Integration

```bash
go get google.golang.org/api
```

- Official Google APIs client library
- Used for **Google Calendar integration**
- Allows staff to sync bookings with Google Calendar
- Provides calendar event creation and deletion

---

### Rate Limiting

```bash
go get golang.org/x/time/rate
```

- Token bucket rate limiter
- Protects API endpoints from abuse
- Different limits for OAuth and API routes

---

### HTTP Retry Logic

```bash
go get github.com/hashicorp/go-retryablehttp
```

- HTTP client with automatic retry logic
- Used for external API calls (42 Intra OAuth)
- Handles transient failures gracefully

```bash
go get github.com/avast/retry-go/v5
```

- Retry mechanism for operations
- Used for database operations and external service calls
- Configurable backoff strategies

---

## go.mod Overview 

```bash
module github.com/IbnBaqqi/book-me

go 1.25.7

require (
	github.com/avast/retry-go/v5 v5.0.0
	github.com/go-playground/validator/v10 v10.30.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/sessions v1.4.0
	github.com/hashicorp/go-retryablehttp v0.7.8
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.11.1
	github.com/wneessen/go-mail v0.7.2
	golang.org/x/oauth2 v0.35.0
	golang.org/x/time v0.14.0
	google.golang.org/api v0.265.0
)
```

---

### Why This Stack?

- **Pure `net/http`** – no heavy frameworks, full control over routing and middleware
- **go-playground/validator** – powerful struct validation with custom rules
- OAuth handled explicitly for learning and control
- PostgreSQL + SQLC for type-safe SQL queries
- Google Calendar API for seamless staff workflow integration
- Rate limiting to prevent API abuse
- Retry logic for resilient external API calls
- Clear separation between:
  - Authentication (OAuth + JWT)
  - Business rules (service layer)
  - Infrastructure (email, DB, OAuth, Google Calendar)

This keeps the project **simple, explicit, and easy to reason about**.

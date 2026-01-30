<div align="center">

# BookMe – Meeting Room Reservation Backend

[![Go Version](https://img.shields.io/github/go-mod/go-version/ibnbaqqi/book-me?style=flat&logo=go&color=00ADD8)](https://github.com/ibnbaqqi/book-me/blob/main/go.mod)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18.1-336791?logo=postgresql)
[![Go Report Card](https://goreportcard.com/badge/github.com/ibnbaqqi/book-me?style=flat)](https://goreportcard.com/report/github.com/ibnbaqqi/book-me)
![Google Calendar](https://img.shields.io/badge/Google_Cal-API-4285F4?style=flat&logo=google-calendar)
![Keycloak](https://img.shields.io/badge/Keycloak-v21+-662222?style=flat&logo=keycloak)

</div>

---
<div align="center">
</div>

# Purpose

Book-me is a modern meeting room booking system that allows students and staff to book meeting rooms at Hive Helsinki.

It supports calendar-based views, role-based access control (students & staff), and 42 Intra OAuth2 authentication.

- **Live WebApp:** [room.hive.fi](https://room.hive.fi) [Hive Login required]
- **Frontend:** [Frontend Source code](https://github.com/danielxfeng/booking_calendar.git)
---

### Basic System Architecture Diagram
![System Architecture](assets/v3BookMe-whiteBg.png)

## Features

- **42 Intra OAuth2 Login**: Secure authentication using Hive Helsinki’s 42 Intranet
- **Smart Booking Logic**: Prevents overlapping reservations and restricts cancellation rights
- **Role-Based Access**:
  - Staff can view who booked each slot and cancel any booking
  - Students can only see availability and cancel their own bookings
- **Calendar API**: Fetches unavailable time slots for specific date ranges
- **Secure JWT Authentication**: Stateless session management using JSON Web Tokens
- **Email Notifications**: Sends confirmations and updates to users
- **Google Calender Integration**: Allow staff to view update without leaving their workflow

---
## Tech Stack
- **Go 1.21+, PostgreSQL with SQLC**
---

## Project Structure

```bash
book-me/
├── cmd/
│   └── api/
│       └── main.go                 # Moved from root (TODO)
├── internal/
│   ├── auth/                       # Logic for JWT/Auth
│   │   ├── auth.go
│   │   └── auth_test.go
│   ├── database/                   # SQLC generated code
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── reservations.sql.go
│   │   ├── rooms.sql.go
│   │   └── users.sql.go
│   ├── email/                      # Email service & templates
│   │   ├── email_service.go
│   │   ├── email_service_test.go
│   │   └── templates/
│   ├── external/                   # Third-party integrations
│   │   └── google/
│   │       ├── auth.go
│   │       ├── calender_service.go
│   │       ├── models.go
│   │       └── token_manager.go
│   ├── handlers/                   # HTTP Handlers (moved from root) (TODO)
│   │   ├── handler_callback.go
│   │   ├── handler_login.go
│   │   ├── handler_reservation_delete.go
│   │   ├── handler_reservations.go
│   │   ├── handler_get_reservations.go
│   │   ├── health.go
│   │   └── json.go                 # JSON helper utilities
│   └── services/                   # Business logic layers
│       ├── users_service.go        # Moved from root
│       └── user_42.go              # Specific business logic
├── sql/                            # Database migrations and queries
│   ├── queries/
│   │   ├── reservations.sql
│   │   ├── rooms.sql
│   │   └── users.sql
│   └── schema/
│       ├── 001_users.sql
│       ├── 002_rooms.sql
│       ├── 003_reservations.sql
│       └── 004_populate_rooms.sql
├── docs/                           # Documentation
│   ├── api_overview.md
│   ├── setup.md
│   └── usage.md
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── sqlc.yaml
├── README.md
└── v3BookMe-whiteBg.png
```

---

## Getting Started

### Requirements

- Go 1.22+, PostgreSQL, 42 Intra client ID/secret

#### Setup Instructions [here](docs/setup.md)

#### API Overview [here](docs/api_overview.md)

#### Library & Dependencies [here](docs/usage.md)

---

### Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

---

MIT License

---
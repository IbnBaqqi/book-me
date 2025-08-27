
# BookMe – Meeting Room Reservation Backend

Book-me is a backend API that allows students and staff to book meeting rooms at Hive Helsinki.
It supports calendar-based views, role-based access control (students & staff), and 42 Intra OAuth2 authentication.

- **Live Preview:** [room.hive.fi](https://room.hive.fi) [Hive Login required]
- **Frontend:** [Frontend Source code](https://github.com/danielxfeng/booking_calendar.git)
---

### Basic System Architecture Diagram
![System Architecture](/v3BookMe-noBgDark.png)

## Features

- **42 Intra OAuth2 Login**: Secure authentication using Hive Helsinki’s 42 Intranet
- **Smart Booking Logic**: Prevents overlapping reservations and restricts cancellation rights
- **Role-Based Access**:
  - Staff can view who booked each slot and cancel any booking
  - Students can only see availability and cancel their own bookings
- **Calendar API**: Fetches unavailable time slots for specific date ranges
- **Secure JWT Authentication**: Stateless session management using JSON Web Tokens
- **Email Notifications**: Sends confirmations and updates to users

---
## Tech Stack
- **Spring Boot (Security, JPA), Java 17+, MySQL, Lombok, MapStruct**
---

## Project Structure

```bash
src/main/java/com/hivestudent/bookme/
├── Auth/               # OAuth2 login logic & Jwt
│ ├── JwtFilter         # Filter for Jwt on every request
│ ├── JwtService        # Jwt logic
│ ├── OAuthController   # 42 OAuth2 endpoints
│ ├── OAuthService      # OAuth2 logic
├── config/             # Security configurations
├── controllers/        # API endpoints
├── dao/                # Data Access Object / JPA Repositories
├── dtos/               # Response/request models
│ ├── ReservedDto       # Response model for /reservation/unavailable
├── entities/           # Room, Reservation, User, etc.
├── exceptions/         # Exception Handler
├── Google/             # Google Auth & Calender integration
├── mapper/             # Mapstruct AutoMapper
└── services/           # Business logic
```

---

## Getting Started

### Requirements

- Java 17+, MySQL, 42 Intra client ID/secret

### Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/IbnBaqqi/book-me.git
cd book-me
```

### 2. MySQL Configuration

Create local database "bookMe"

In `src/main/resources/application-dev.yaml`, add:

- `spring.datasource.url`=jdbc:mysql://localhost:3306/bookMe
- `spring.datasource.username`=your_mysql_username (most likely "root")
- `spring.datasource.password`=your_mysql_password

### 3. Configure Environment Variables
- Rename the ``.env.yaml.example`` file to ``.env.yaml``.
- Update the following environment variables inside .env:

#### JWT_SECRET

Generate a secure random key using openssl or any secure generator tool:
```bash
openssl rand -base64 32
```

#### 42 CREDENTIALS

1. Generate a new API application on the [42 intranet](https://profile.intra.42.fr/oauth/applications/new)
2. In the field Redirect URI add: http://localhost:8080/oauth/callback
3. From the available scopes, choose "Access the user public data" and then proceed to submit.

- Environment variables:
- `CLIENT_ID`: 42 API client ID
- `SECRET`: 42 API client secret
- `REDIRECT_URI`: http://localhost:8080/oauth/callback
- `OAUTH_AUTH_URI`: https://api.intra.42.fr/oauth/authorize
- `OAUTH_TOKEN_URI`: https://api.intra.42.fr/oauth/token
- `JWT_SECRET`: YOUR JWT_SECRET
- `REDIRECT_TOKEN_URI`: http://localhost:8080/?token=

---
Then run the app.
For Windows, use mvnw.cmd:

```bash
./mvnw spring-boot:run
```

---

#### API Overview [here](docs/api_overview.md)

---

### Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

---

### License

MIT License

---

### 💡 Todo
- [x] Use Docker
- [ ] Add Swagger
- [x] System Architecture Diagram
- [ ] Sequence Diagram
- [x] Email notifications
- [x] Google Calendar Integration

---

# 📅 BookMe – Meeting Room Reservation Backend

Book-me is a backend API that allows students and staff to book meeting rooms at Hive Helsinki.
It supports calendar-based views, role-based access control (students & staff), and 42 Intra OAuth2 authentication.

- **Live Preview:** [booking-calendar-chi.vercel.app](https://booking-calendar-chi.vercel.app)
- **Link to Frontend:** [https://github.com/danielxfeng/booking_calendar.git](https://github.com/danielxfeng/booking_calendar.git)
- 
---

## 🚀 Features

- 🔒 42 Intra OAuth2 login
- 🏫 Room creation & management (admin only)
- 📆 Reservation of time slots for available rooms
- 👨‍🏫 Staff can see who booked a slot
- 👨‍🎓 Students can only see availability
- ❌ Cancel bookings (only by creator or staff)
- 📊 Calendar view (week/day/month)
- 🔄 Unavailable time slots API
- 📤 RESTful JSON API ready for frontend integration

---

## 🛠️ Tech Stack

- **Java 17+**
- **Spring Boot 3**
- **Spring Security**
- **JPA + MySQL**
- **Lombok, MapStruct**
- **REST API**

---

## 📦 Project Structure

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
├── mapper/             # Mapstruct AutoMapper
└── services/           # Reservation logic
```

---

## ✅ Getting Started

### Requirements

- Java 17+
- MySQL
- 42 Intra client ID/secret

### Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/IbnBaqqi/book-me.git
cd book-me
```

### 2. MySQL Configuration

Create local database "bookMe"

In `src/main/resources/application-dev.yaml`, add:

```properties
spring.datasource.url=jdbc:mysql://localhost:3306/bookMe
spring.datasource.username=your_mysql_username (most likely "root")
spring.datasource.password=your_mysql_password
  ```

### 3. Configure Environment Variables
- Rename the ``.env.yaml.example`` file to ``.env.yaml``.
- Update the following environment variables inside .env:

#### JWT_SECRET

Generate a secure random key using:

```bash
openssl rand -base64 32
```

If ``openssl`` is not available, go to [generate-random.org](https://generate-random.org), click on **Strings > API Tokens**, and generate a secure token.

#### 42 CREDENTIALS

1. Generate a new API application on the [42 intranet](https://profile.intra.42.fr/oauth/applications/new)
2. In the field Redirect URI add: http://localhost:8080/oauth/callback
3. From the available scopes, choose "Access the user public data" and then proceed to submit.
4. Save the credentials, you will need them in the next step.


```properties
FORTY_TWO:
    CLIENT_ID:YOUR 42 CLIENT_ID
    SECRET: YOUR 42_SECRET
    REDIRECT_URI: http://localhost:8080/oauth/callback
    OAUTH_AUTH_URI: https://api.intra.42.fr/oauth/authorize
    OAUTH_TOKEN_URI: https://api.intra.42.fr/oauth/token
JWT_SECRET: YOUR JWT_SECRET
REDIRECT_TOKEN_URI: http://localhost:8080/?token=
```

Then run the app:

```bash
./mvnw spring-boot:run
```

If you're on Windows:

```bash
mvnw.cmd spring-boot:run
```

Once running, the application will be available at:

```arduino
http://localhost:8080
```

---

## 📚 API Overview

### 🔐 Auth

| Method | Endpoint    | Description               |
|--------|-------------|---------------------------|
| GET    | `/callback` | 42 OAuth2 redirect URI    |
| GET    | `/login`    | Authentication entrypoint |

---


### 📆 Reservations

| Method | Endpoint                    | Description                           |
|--------|-----------------------------|---------------------------------------|
| POST   | `/reservations`             | Create reservation                    |
| DELETE | `/reservations/{id}`        | Cancel if you're staff or owner       |
| GET    | `/reservations/me`          | Get user's own bookings               |
| GET    | `/reservations/unavailable` | Get all booked slots for a week range |

Query params for /unavailable:

```
?start=2025-06-20&end=2025-06-27
```

---

## 👥 Roles & Access

| Role   | Permissions                                               |
|--------|-----------------------------------------------------------|
| Student | Book, view available slots                               |
| Staff   | Book, cancel anyone’s slot, view who booked              |

---

## 🔧 Example JSON Response for `/reservation/unavailable`

```json
[
  {
    "roomId": 1,
    "roomName": "Small Room",
    "slots": [
      {
        "start": "2025-06-24T10:00",
        "end": "2025-06-24T11:30",
        "bookedBy": "Intra UserName"
      }
    ]
  }
]
```

## 🙌 Contributing

Contributions are welcome!

1. Fork the project
2. Create a new branch
3. Commit your changes
4. Submit a PR

---

## 📄 License

MIT — free to use and modify.

---

## 💡 Todo
- [ ] Use Docker for Ease
- [ ] Add Swagger
- [ ] Email/Discord notifications
- [ ] Recurrent bookings (weekly meetings)
- [ ] Google Calender Integration

---

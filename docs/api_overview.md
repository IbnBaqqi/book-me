## API Overview

### Auth

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

## Roles & Access

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
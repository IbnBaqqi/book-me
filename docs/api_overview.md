## API Overview

Authentication Flow

1. User clicks "Login" → Redirected to /api/oauth/login
2. System redirects to OAuth provider (42 Intra)
3. User authorizes the application
4. OAuth provider redirects to /oauth/callback
5. System exchanges code for access token
6. System fetches user info and creates/updates user in database
7. System generates JWT token
8. JWT token returned to client
9. Client includes JWT in Authorization: Bearer <token> header for protected routes

## API Endpoints
### Authentication

| Method | Endpoint              | Description                 | Auth Required |
|------|-----------------------|-----------------------------|---------------|
| GET  | /api/oauth/login      | Initiate OAuth login        | No            |
| GET  | /oauth/callback       | OAuth callback handler      | No            |

### Reservations

| Method | Endpoint                         | Description                         | Auth Required |
|------|----------------------------------|-------------------------------------|---------------|
| POST | /reservation                     | Create a new reservation            | Yes           |
| GET  | /reservation                     | Get unavailable time slots          | Optional      |
| DELETE | /reservation/{id}              | Cancel a reservation                | Yes           |

### Health Check

| Method | Endpoint        | Description             | Auth Required |
|------|-----------------|-------------------------|---------------|
| GET  | /api/healthz    | Health check endpoint   | No            |

---

## API Examples 📝

### Create a Reservation

```bash
curl -X POST http://localhost:8080/reservation \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roomId": 1,
    "startTime": "2025-01-28T14:00:00Z",
    "endTime": "2025-01-28T16:00:00Z"
  }'
```

**Response**

```json
{
  "id": 123,
  "roomId": 1,
  "startTime": "2025-01-28T14:00:00Z",
  "endTime": "2025-01-28T16:00:00Z",
  "createdBy": {
    "id": 42,
    "name": "John Doe"
  }
}
```

---

### Get Unavailable Slots

```bash
curl "http://localhost:8080/reservation?start=2025-01-28&end=2025-02-01"
```

**Response**

```json
[
  {
    "roomId": 1,
    "roomName": "Conference Room A",
    "slots": [
      {
        "id": 123,
        "startTime": "2025-01-28T14:00:00Z",
        "endTime": "2025-01-28T16:00:00Z",
        "bookedBy": "John Doe"
      }
    ]
  }
]
```

---

### Cancel a Reservation

```bash
curl -X DELETE http://localhost:8080/reservation/123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response**

```
204 No Content
```

---

## Business Rules 📐

### Reservation Rules

#### Time Validation

- Cannot book past times
- End time must be after start time
- Maximum duration:
  - **Students**: 4 hours
  - **Staff**: Unlimited

---

#### Overlap Prevention

- System checks for conflicting reservations
- Returns **400 Bad Request** if slot is already booked

---

#### Authorization

- Users can only cancel **their own** reservations
- Staff members can cancel **any** reservation

---

#### Privacy

Booking details are visible to:
- The person who made the booking
- Staff members

Other users only see time slots marked as **"booked"** without personal details.

---

## Email Notifications 📧

The system automatically sends emails for:

- ✅ **Booking Confirmation** – sent immediately after a successful reservation
- 🚫 **Cancellation Notice** – sent when a reservation is cancelled

### Gmail SMTP Setup

1. Enable **2-Factor Authentication** on your Google account
2. Generate an **App Password**  
   https://myaccount.google.com/apppasswords  
3. Select **Mail** and your device
4. Copy the generated 16-character password
5. Use this password as:

```bash
SMTP_PASSWORD=your-gmail-app-password
```

---

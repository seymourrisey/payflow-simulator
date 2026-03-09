## Payflow Simulator Backend

- Go 1.21+
- PostgreSQL (or a Supabase project)

## Setup

1. Clone the repository and navigate to the backend directory.

2. Create a `.env` file in the project root (or inside `backend/`):
```env
APP_PORT=8080
APP_ENV=development
DATABASE_URL=your-database-url
JWT_SECRET=your-secret-key
JWT_EXPIRY_HOURS=24
WEBHOOK_TIMEOUT_SECONDS=10
WEBHOOK_MAX_RETRIES=3
ALLOW_ORIGINS=http://localhost:5173
```

3. Run the database migration by executing `backend/db/migrations/payflow-database.sql` on your PostgreSQL instance.

4. Start the backend:

```bash
go mod tidy
```

```bash
cd cmd/api
go run main.go
```

---

## API Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Login and get JWT token |
| POST | `/api/auth/logout` | Logout (client-side token removal) |
| GET | `/api/wallet` | Get wallet balance |
| POST | `/api/wallet/topup` | Initiate a top-up request |
| POST | `/api/payment/qr` | Generate a payment QR code |
| POST | `/api/payment/pay` | Process a QR payment |
| GET | `/api/transactions` | Get transaction history |
| GET | `/api/webhooks` | Get webhook delivery logs |
| GET | `/api/webhooks/stats` | Get webhook delivery statistics |
| GET | `/api/webhooks/merchants` | List merchants |
| POST | `/webhook/receive` | Built-in webhook receiver (for testing) |

Protected routes require the `Authorization: Bearer <token>` header.

---

## Transaction IDs

The system uses human-readable prefixed IDs:

| Entity | Format |
|--------|--------|
| User | `USR-A1B2C3D4E5F6` |
| Wallet | `WLT-A1B2C3D4E5F6` |
| Merchant | `MRC-A1B2C3D4E5F6` |
| Transaction | `TXN-20260307-A1B2C3D4` |
| Top-up Request | `TUP-20260307-A1B2C3D4` |

---

## License

MIT

# Payflow Simulator - 

A full-stack payment flow simulator built to demonstrate how a digital wallet and payment gateway system works end-to-end, including QR-based payments, top-up flows, transaction history, merchant webhooks, and webhook delivery monitoring.

---

## Tech Stack

**Backend**
- Go 1.25 with [Gin](https://github.com/gin-gonic/gin)
- PostgreSQL via [pgx/v5](https://github.com/jackc/pgx)
- JWT authentication (golang-jwt/jwt v4)
- Docker (multi-stage build)

**Frontend**
- React 18 + Vite
- Tailwind CSS
- Zustand (state management)
- Axios, React Router v6, qrcode.react, jsQR

---

## Features

- User registration and login with JWT-based authentication
- Digital wallet per user (IDR currency, balance cannot go negative)
- Top-up via simulated payment channels (Bank Transfer, Virtual Account)
- QR code payment generation and scanning
- Payment processing with idempotency key support
- Transaction history with pagination (PAYMENT, TOPUP, TRANSFER)
- Merchant management with webhook URL configuration
- Automatic webhook dispatch with retry mechanism (up to 3 retries)
- Built-in webhook receiver endpoint for local testing
- Webhook delivery logs and stats dashboard

---

### Demo deploy 
back4app & vercel

# Yaxshilik Izlab — Backend

Go + Gin + Postgres backend, Click va Payme to'lov tizimlari bilan to'liq integratsiya, admin panel bilan birga.

## Ishga tushirish

### 1. Postgres tayyorlash

`.env` da `PG_URL` to'g'ri ekanini tekshiring. Lokal sinov uchun:

```bash
docker run -d --name iu-pg -p 5432:5432 \
  -e POSTGRES_PASSWORD=1111 \
  -e POSTGRES_DB=ptest \
  postgres:16-alpine
```

Va `.env` da `PG_URL=postgres://postgres:1111@localhost:5432/ptest` qiling.

### 2. Backend ishga tushirish

```bash
cd backend
go run .
```

Birinchi ishga tushirishda:
- Barcha jadvallar avtomatik yaratiladi (`migrations.sql`)
- `.env` dagi `ADMIN_EMAIL` / `ADMIN_PASSWORD` bilan admin yaratiladi

### 3. URL'lar

| Manzil | Nima |
|--------|------|
| `http://localhost:8080/` | Asosiy sayt (public, index.html) |
| `http://localhost:8080/admin` | Admin panel |
| `http://localhost:8080/api/people` | Bemorlar (public) |
| `http://localhost:8080/api/click/webhook` | Click → bizga |
| `http://localhost:8080/api/payme` | Payme → bizga (JSON-RPC) |

## API

### Public

- `GET  /api/people` — faol bemorlar
- `GET  /api/people/:id` — bitta bemor
- `GET  /api/stats/public` — umumiy statistika
- `POST /api/click/create` — Click to'lov boshlash → `redirect_url`
- `POST /api/payme/create` — Payme to'lov boshlash → `redirect_url`
- `POST /api/click/webhook` — Click webhook (Prepare/Complete, MD5 imzo)
- `POST /api/payme` — Payme JSON-RPC (6 metod, Basic Auth)

### Admin (JWT)

- `POST   /api/admin/login` — `{ email, password }` → `{ token, admin }`
- `GET    /api/admin/me`
- `GET    /api/admin/stats`
- `GET    /api/admin/people`
- `POST   /api/admin/people`
- `PUT    /api/admin/people/:id`
- `DELETE /api/admin/people/:id`
- `GET    /api/admin/donations?status=&provider=&person_id=&limit=&offset=`
- `POST   /api/admin/upload` (multipart `file`)

## Click va Payme sozlash

### Click kabinetiga kirib webhook URL ko'rsating:
```
https://your-domain.uz/api/click/webhook
```

### Payme kabinetiga kirib webhook URL ko'rsating:
```
https://your-domain.uz/api/payme
```

`Authorization` headerida `Basic Base64(Paycom:KEY)` kelishi kerak — bizning kod buni avtomatik tekshiradi.

## Xavfsizlik

- `.env` git'ga commit qilinmaydi (`.gitignore` da)
- Click `secret_key` MD5 imzoni tekshirish uchun ishlatiladi
- Payme `secret_key` Basic Auth tekshiruvi uchun ishlatiladi
- JWT 72 soat amal qiladi
- Webhook'lar idempotent — bir tranzaksiya ikki marta hisoblanmaydi
- Hech qachon Click `amount` qiymatiga ishonmaymiz — DB dagi summa bilan solishtiramiz
- Payme ham xuddi shunday — `account.order_id` orqali DB dagi donation topiladi

## Qo'shimcha xavfsizlik sozlamalari (.env)

Quyidagi kalitlar ixtiyoriy, lekin production uchun tavsiya etiladi:

```bash
# CORS — ruxsat etilgan frontend domenlari (vergul bilan).
# Bo'sh + GIN_MODE=release => tashqi domenlarga umuman ruxsat yo'q (same-origin ishlaydi).
# Bo'sh + debug => barcha domenlarga ruxsat (faqat lokal test uchun).
CORS_ORIGINS=https://your-domain.uz

# Webhook IP whitelist — Click/Payme serverlari IP/CIDR (vergul bilan). Bo'sh => o'chiq.
WEBHOOK_ALLOW_IPS=213.230.106.115,213.230.106.116,185.178.49.0/24

# Reverse proxy (nginx/caddy) ortida bo'lsa, haqiqiy mijoz IP'sini X-Forwarded-For dan olish
TRUST_PROXY=true

# Payme uchun alohida return URL (bo'sh => CLICK_RETURN_URL ishlatiladi)
PAYME_RETURN_URL=https://your-domain.uz/?paid=1
```

**Avtomatik himoyalar (kodga o'rnatilgan):**
- To'lov tasdiqlash atomik (race-safe) — pul ikki marta hisoblanmaydi
- `people.raised` cheklanmaydi => refund to'g'ri ishlaydi
- Imzo va secret'lar **constant-time** solishtiriladi (timing attack himoyasi)
- Admin login: **10 urinish / 5 daqiqa / IP** (brute-force himoyasi)
- Donate endpointlar: **20 so'rov / daqiqa / IP**
- Upload: faqat rasm (jpg/png/webp/gif), maks **8 MB**, MIME tekshiriladi, fayl nomi `crypto/rand`
- Security headers (nosniff, X-Frame-Options va h.k.)
- `release` rejimida zaif JWT_SECRET / default ADMIN_PASSWORD bilan ishga tushmaydi
- Muhim hodisalar (to'lov, refund, login) `audit_log` jadvaliga yoziladi

## Production'ga deploy

- `.env` da `GIN_MODE=release` qiling
- `JWT_SECRET` ni uzunroq tasodifiy stringga almashtiring (`openssl rand -hex 32`)
- `ADMIN_PASSWORD` ni mustahkamiga o'zgartiring
- HTTPS shart (Click ham, Payme ham faqat HTTPS bilan ishlaydi)
- `WEBHOOK_ALLOW_IPS` ga Click/Payme IP'larini qo'shing (yoki nginx orqali whitelist):
  ```
  213.230.106.115
  213.230.106.116
  ```

## Loyiha tuzilishi

```
yaxshilik_izlab/
├── index.html              # Public sayt
├── admin/index.html        # Admin panel (SPA)
├── .env                    # Kalitlar (git'ga tushmaydi)
├── backend/
│   ├── main.go             # Routes
│   ├── config.go           # .env loader
│   ├── db.go               # pgx pool + migrations
│   ├── migrations.sql      # Postgres schema
│   ├── models.go           # Person, Donation, Admin struct'lari
│   ├── middleware.go       # JWT verify
│   ├── auth.go             # Login, JWT, admin bootstrap
│   ├── people.go           # CRUD
│   ├── click.go            # Click webhook + create
│   ├── payme.go            # Payme JSON-RPC + create
│   ├── transactions.go     # Donations ro'yxati
│   ├── stats.go            # Dashboard
│   └── upload.go           # MinIO upload
```

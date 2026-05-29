# Click To'lov Tizimi — To'liq Hujjat

Click — O'zbekistondagi eng yirik to'lov tizimlaridan biri. Bu hujjat backend integratsiyasi uchun mo'ljallangan.

---

## Mundarija

1. [Asosiy Tushunchalar](#asosiy-tushunchalar)
2. [Hisob Ma'lumotlari (Credentials)](#hisob-malumotlari)
3. [Imzo (Signature) Yaratish](#imzo-yaratish)
4. [Shop API (Invoice yaratish)](#shop-api)
5. [Merchant API (Webhook)](#merchant-api-webhook)
6. [Standart Oqim (Flow)](#standart-oqim)
7. [Xatolik Kodlari](#xatolik-kodlari)
8. [Test Kartalar](#test-kartalar)
9. [Xavfsizlik](#xavfsizlik)
10. [Tez-tez Uchraydigan Savollar](#faq)

---

## Asosiy Tushunchalar

| Termin | Ma'nosi |
|--------|---------|
| `service_id` | Click tomonidan beriladigan xizmat ID si |
| `merchant_id` | Sizning Click hisobingiz ID si |
| `merchant_user_id` | Click kabinetingizdagi user ID |
| `secret_key` | Imzo (signature) yaratish uchun maxfiy kalit |
| `merchant_trans_id` | **SIZNING** tranzaksiya ID ingiz (DB dagi) |
| `click_trans_id` | Click tomonidan beriladigan tranzaksiya ID |
| `amount` | To'lov summasi (so'm da, masalan: `50000.00`) |
| `action` | `0` = Prepare, `1` = Complete |

---

## Hisob Ma'lumotlari

Click bilan shartnoma tuzgandan keyin quyidagilar beriladi:

```
SERVICE_ID       = 12345
MERCHANT_ID      = 67890
MERCHANT_USER_ID = 11111
SECRET_KEY       = "your_secret_key_here"
```

Bularni **`.env`** faylida saqlang, hech qachon kodga yozmang!

```bash
# .env
CLICK_SERVICE_ID=12345
CLICK_MERCHANT_ID=67890
CLICK_MERCHANT_USER_ID=11111
CLICK_SECRET_KEY=your_secret_key_here
```

---

## Imzo Yaratish

Click har bir so'rovda **MD5 hash** orqali imzo yuboradi. Sizning vazifangiz — uni qayta hisoblab, mos kelishini tekshirish.

### Prepare uchun signature:

```
md5(click_trans_id + service_id + secret_key + merchant_trans_id + amount + action + sign_time)
```

### Complete uchun signature:

```
md5(click_trans_id + service_id + secret_key + merchant_trans_id + merchant_prepare_id + amount + action + sign_time)
```

### Python misol:

```python
import hashlib

def generate_signature(
    click_trans_id: str,
    service_id: str,
    secret_key: str,
    merchant_trans_id: str,
    amount: str,
    action: str,
    sign_time: str,
    merchant_prepare_id: str = ""
) -> str:
    if action == "0":  # Prepare
        raw = f"{click_trans_id}{service_id}{secret_key}{merchant_trans_id}{amount}{action}{sign_time}"
    else:  # Complete
        raw = f"{click_trans_id}{service_id}{secret_key}{merchant_trans_id}{merchant_prepare_id}{amount}{action}{sign_time}"

    return hashlib.md5(raw.encode()).hexdigest()
```

### Shop API uchun Auth header:

```python
import hashlib
from datetime import datetime

def make_auth_header(merchant_user_id: str, secret_key: str) -> str:
    timestamp = str(int(datetime.now().timestamp()))
    digest = hashlib.sha1(f"{timestamp}{secret_key}".encode()).hexdigest()
    return f"{merchant_user_id}:{digest}:{timestamp}"

# Header:
# Auth: 11111:abc123def456...:1716451200
```

---

## Shop API

Shop API foydalanuvchini to'lov sahifasiga yo'naltirish uchun ishlatiladi.

**Base URL:** `https://api.click.uz/v2/merchant/`

### 1. Invoice yaratish

```bash
curl -X POST 'https://api.click.uz/v2/merchant/invoice/create' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'Auth: 11111:abc123def456:1716451200' \
  -d '{
    "service_id": 12345,
    "amount": 100000.00,
    "phone_number": "998901234567",
    "merchant_trans_id": "TXN-2026-0001"
  }'
```

**Javob:**
```json
{
  "error_code": 0,
  "error_note": "Success",
  "invoice_id": 987654321
}
```

### 2. Invoice holatini tekshirish

```bash
curl -X GET 'https://api.click.uz/v2/merchant/invoice/status/12345/987654321' \
  -H 'Auth: 11111:abc123def456:1716451200'
```

**Javob:**
```json
{
  "error_code": 0,
  "error_note": "Success",
  "invoice_status": 2,
  "invoice_status_note": "Paid"
}
```

| invoice_status | Ma'nosi |
|----------------|---------|
| 0 | Yaratilgan, kutilmoqda |
| 1 | Bekor qilingan |
| 2 | To'langan |

### 3. Tranzaksiya holatini tekshirish

```bash
curl -X GET 'https://api.click.uz/v2/merchant/payment/status/12345/TXN-2026-0001' \
  -H 'Auth: 11111:abc123def456:1716451200'
```

### 4. Pulni qaytarish (Refund)

```bash
curl -X DELETE 'https://api.click.uz/v2/merchant/payment/reversal/12345/PAYMENT_ID' \
  -H 'Auth: 11111:abc123def456:1716451200'
```

### 5. To'g'ridan-to'g'ri to'lov havolasi (Redirect URL)

Eng oddiy usul — userni shu URL ga yo'naltirish:

```
https://my.click.uz/services/pay?service_id=12345&merchant_id=67890&amount=100000&transaction_param=TXN-2026-0001&return_url=https://yoursite.uz/success
```

---

## Merchant API (Webhook)

Click sizning serveringizga **2 ta so'rov** yuboradi:

1. **Prepare** (`action=0`) — to'lovni tasdiqlash
2. **Complete** (`action=1`) — to'lovni yakunlash

### Endpoint sozlash

Click kabinetida 1 ta URL ko'rsatasiz:
```
https://yoursite.uz/api/click/webhook
```

Click bu URL ga `POST` so'rov yuboradi `application/x-www-form-urlencoded` formatida.

### Kiruvchi parametrlar:

```
click_trans_id       = 123456789
service_id           = 12345
click_paydoc_id      = 987654
merchant_trans_id    = TXN-2026-0001
amount               = 100000.00
action               = 0  (yoki 1)
error                = 0
error_note           = Success
sign_time            = 2026-05-23 10:30:45
sign_string          = md5_hash_qiymati
merchant_prepare_id  = (faqat Complete uchun)
```

### Prepare so'rovi (curl misol — Click yuboradigan ko'rinish):

```bash
curl -X POST 'https://yoursite.uz/api/click/webhook' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'click_trans_id=123456789' \
  -d 'service_id=12345' \
  -d 'click_paydoc_id=987654' \
  -d 'merchant_trans_id=TXN-2026-0001' \
  -d 'amount=100000.00' \
  -d 'action=0' \
  -d 'error=0' \
  -d 'error_note=' \
  -d 'sign_time=2026-05-23 10:30:45' \
  -d 'sign_string=a1b2c3d4e5f6...'
```

### Sizning Prepare javobingiz:

```json
{
  "click_trans_id": 123456789,
  "merchant_trans_id": "TXN-2026-0001",
  "merchant_prepare_id": 555,
  "error": 0,
  "error_note": "Success"
}
```

### Complete so'rovi:

```bash
curl -X POST 'https://yoursite.uz/api/click/webhook' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'click_trans_id=123456789' \
  -d 'service_id=12345' \
  -d 'click_paydoc_id=987654' \
  -d 'merchant_trans_id=TXN-2026-0001' \
  -d 'merchant_prepare_id=555' \
  -d 'amount=100000.00' \
  -d 'action=1' \
  -d 'error=0' \
  -d 'error_note=' \
  -d 'sign_time=2026-05-23 10:31:00' \
  -d 'sign_string=f6e5d4c3b2a1...'
```

### Sizning Complete javobingiz:

```json
{
  "click_trans_id": 123456789,
  "merchant_trans_id": "TXN-2026-0001",
  "merchant_confirm_id": 777,
  "error": 0,
  "error_note": "Success"
}
```

---

## Standart Oqim

```
┌──────────┐                ┌──────────┐               ┌──────────┐
│  USER    │                │ BACKEND  │               │  CLICK   │
└────┬─────┘                └────┬─────┘               └────┬─────┘
     │                           │                          │
     │ 1. "100 000 to'ldirish"   │                          │
     ├──────────────────────────>│                          │
     │                           │                          │
     │                           │ 2. DB ga yozish:         │
     │                           │   tx_id, amount=100000,  │
     │                           │   status=pending         │
     │                           │                          │
     │                           │ 3. Invoice yaratish      │
     │                           ├─────────────────────────>│
     │                           │<─────────────────────────┤
     │                           │   invoice_id qaytadi     │
     │                           │                          │
     │ 4. Click sahifasiga       │                          │
     │    redirect               │                          │
     │<──────────────────────────┤                          │
     │                                                      │
     │ 5. Karta ma'lumotlarini kiritish                     │
     ├─────────────────────────────────────────────────────>│
     │                                                      │
     │                           │ 6. Prepare webhook       │
     │                           │<─────────────────────────┤
     │                           │   (amount tekshiriladi)  │
     │                           ├─────────────────────────>│
     │                           │   error=0, prepare_id    │
     │                           │                          │
     │                           │ 7. Complete webhook      │
     │                           │<─────────────────────────┤
     │                           │                          │
     │                           │ 8. DB: status=paid       │
     │                           │    user.balance += 100K  │
     │                           ├─────────────────────────>│
     │                           │   error=0                │
     │                           │                          │
     │ 9. Success page           │                          │
     │<──────────────────────────┤                          │
```

---

## To'liq Python (FastAPI) Misol

```python
import hashlib
from datetime import datetime
from fastapi import FastAPI, Form, HTTPException
from sqlalchemy.orm import Session

app = FastAPI()

CLICK_SERVICE_ID = "12345"
CLICK_SECRET_KEY = "your_secret_key"

# --- 1. Imzo tekshirish ---
def verify_signature(data: dict) -> bool:
    action = data["action"]
    if action == "0":  # Prepare
        raw = (
            f"{data['click_trans_id']}"
            f"{data['service_id']}"
            f"{CLICK_SECRET_KEY}"
            f"{data['merchant_trans_id']}"
            f"{data['amount']}"
            f"{action}"
            f"{data['sign_time']}"
        )
    else:  # Complete
        raw = (
            f"{data['click_trans_id']}"
            f"{data['service_id']}"
            f"{CLICK_SECRET_KEY}"
            f"{data['merchant_trans_id']}"
            f"{data['merchant_prepare_id']}"
            f"{data['amount']}"
            f"{action}"
            f"{data['sign_time']}"
        )

    expected = hashlib.md5(raw.encode()).hexdigest()
    return expected == data["sign_string"]


# --- 2. Webhook endpoint ---
@app.post("/api/click/webhook")
async def click_webhook(
    click_trans_id: str = Form(...),
    service_id: str = Form(...),
    merchant_trans_id: str = Form(...),
    amount: str = Form(...),
    action: str = Form(...),
    sign_time: str = Form(...),
    sign_string: str = Form(...),
    merchant_prepare_id: str = Form(None),
    error: str = Form("0"),
    db: Session = ...,
):
    data = {
        "click_trans_id": click_trans_id,
        "service_id": service_id,
        "merchant_trans_id": merchant_trans_id,
        "amount": amount,
        "action": action,
        "sign_time": sign_time,
        "sign_string": sign_string,
        "merchant_prepare_id": merchant_prepare_id or "",
    }

    # Imzo tekshirish
    if not verify_signature(data):
        return {"error": -1, "error_note": "Signature check failed"}

    # Tranzaksiyani DB dan olish
    tx = db.query(Transaction).filter_by(id=merchant_trans_id).first()
    if not tx:
        return {"error": -5, "error_note": "Transaction not found"}

    # Summani solishtirish
    if float(tx.amount) != float(amount):
        return {"error": -2, "error_note": "Incorrect amount"}

    # PREPARE
    if action == "0":
        if tx.status != "pending":
            return {"error": -4, "error_note": "Already processed"}

        tx.click_trans_id = click_trans_id
        tx.status = "prepared"
        db.commit()

        return {
            "click_trans_id": click_trans_id,
            "merchant_trans_id": merchant_trans_id,
            "merchant_prepare_id": tx.id,
            "error": 0,
            "error_note": "Success",
        }

    # COMPLETE
    elif action == "1":
        if tx.status == "paid":
            return {"error": -4, "error_note": "Already paid"}

        if tx.status != "prepared":
            return {"error": -6, "error_note": "Not prepared"}

        # Agar Click xatolik bilan kelsa
        if error != "0":
            tx.status = "cancelled"
            db.commit()
            return {"error": int(error), "error_note": "Cancelled"}

        # Muvaffaqiyatli to'lov
        tx.status = "paid"
        tx.user.balance += tx.amount
        db.commit()

        return {
            "click_trans_id": click_trans_id,
            "merchant_trans_id": merchant_trans_id,
            "merchant_confirm_id": tx.id,
            "error": 0,
            "error_note": "Success",
        }


# --- 3. Invoice yaratish (foydalanuvchi tomonidan chaqiriladi) ---
@app.post("/api/payment/create")
async def create_payment(amount: float, user_id: int, db: Session = ...):
    if amount < 1000 or amount > 10_000_000:
        raise HTTPException(400, "Summa 1 000 dan 10 000 000 gacha bo'lishi kerak")

    # DB ga tranzaksiya yozish
    tx = Transaction(
        user_id=user_id,
        amount=amount,
        status="pending",
    )
    db.add(tx)
    db.commit()

    # Click redirect URL
    return {
        "payment_url": (
            f"https://my.click.uz/services/pay"
            f"?service_id={CLICK_SERVICE_ID}"
            f"&merchant_id=67890"
            f"&amount={amount}"
            f"&transaction_param={tx.id}"
            f"&return_url=https://yoursite.uz/success"
        )
    }
```

---

## Xatolik Kodlari

Sizning webhook javobingizdagi `error` maydoni:

| Kod | Ma'nosi | Qachon ishlatiladi |
|-----|---------|---------------------|
| `0` | Success | Hammasi yaxshi |
| `-1` | SIGN CHECK FAILED | Imzo noto'g'ri |
| `-2` | Incorrect parameter amount | Summa mos kelmadi |
| `-3` | Action not found | `action` noto'g'ri |
| `-4` | Already paid | Allaqachon to'langan |
| `-5` | User does not exist | User topilmadi |
| `-6` | Transaction does not exist | Tranzaksiya yo'q |
| `-7` | Failed to update user | DB xatolik |
| `-8` | Error in request from click | Parametrlarda xato |
| `-9` | Transaction cancelled | Bekor qilingan |

Click yuboradigan `error` qiymatlari:

| Kod | Ma'nosi |
|-----|---------|
| `0` | Muvaffaqiyatli |
| `-9` | Tranzaksiya bekor qilindi (user to'lashdan voz kechdi) |

---

## Test Kartalar

Click sandbox muhitida:

```
Karta raqami:   8600 4954 7331 6478
Amal qilish:    03/99
SMS kod:        12345
```

Sandbox URL:
```
https://my.click.uz/services/pay  →  https://test.click.uz/services/pay
```

---

## Xavfsizlik (MUHIM!)

### 1. HTTPS shart
Click webhook faqat HTTPS bilan ishlaydi. HTTP da test qilish uchun `ngrok` ishlating:
```bash
ngrok http 8000
```

### 2. IP whitelist (qo'shimcha himoya)
Click serverlari faqat shu IP lardan keladi:
```
213.230.106.115
213.230.106.116
```

Nginx misoli:
```nginx
location /api/click/webhook {
    allow 213.230.106.115;
    allow 213.230.106.116;
    deny all;

    proxy_pass http://backend;
}
```

### 3. Imzoni MAJBURIY tekshiring
```python
if not verify_signature(data):
    return {"error": -1, "error_note": "SIGN CHECK FAILED"}
```

### 4. Idempotency
Bir tranzaksiya 2 marta hisoblanmasligi kerak:
```python
if tx.status == "paid":
    return {"error": -4, "error_note": "Already paid"}
```

### 5. Click yuborgan `amount` ga ishonmang
Har doim DB dagi `tx.amount` bilan solishtiring.

### 6. `secret_key` ni hech qachon log qilmang
```python
# YOMON ❌
logger.info(f"Request data: {data}")  # secret_key bor

# YAXSHI ✅
safe_data = {k: v for k, v in data.items() if k != "secret_key"}
logger.info(f"Request data: {safe_data}")
```

---

## DB Schema Tavsiya

```sql
CREATE TABLE transactions (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL REFERENCES users(id),
    amount              DECIMAL(15, 2) NOT NULL,
    status              VARCHAR(20) DEFAULT 'pending',
    -- pending | prepared | paid | cancelled | failed

    click_trans_id      VARCHAR(50),
    click_paydoc_id     VARCHAR(50),
    merchant_prepare_id VARCHAR(50),

    created_at          TIMESTAMP DEFAULT NOW(),
    paid_at             TIMESTAMP,

    INDEX idx_status (status),
    INDEX idx_user (user_id)
);
```

---

## FAQ

### S: Summa qayerdan keladi — Click dan mi yoki backend dan?
**J:** **Backend dan.** Frontend → Backend (DB ga yozadi) → Click ga uzatadi. Click webhook ga shu summani qaytarib yuboradi, lekin siz **DB dagi qiymat**ga ishonasiz.

### S: User istalgan summani kirita oladimi?
**J:** Ha. Faqat min/max limit qo'ying (masalan 1 000 - 10 000 000 so'm).

### S: Click `tiyin` da ishlaydimi?
**J:** Yo'q, Click `so'm` da ishlaydi (`100000.00` = 100 000 so'm). Lekin **Payme** `tiyin` da ishlaydi — adashtirmang.

### S: Webhook 2 marta kelsa nima bo'ladi?
**J:** Sizning kodingiz `status` ni tekshirib, ikkinchi marta `-4 Already paid` qaytarishi kerak. Bu Click uchun OK.

### S: User to'lamasdan ketib qolsa?
**J:** Click hech qanday webhook yubormaydi. Tranzaksiya `pending` da qoladi. Cron job orqali 24 soatdan keyin `cancelled` qiling.

### S: Refund qilish mumkinmi?
**J:** Ha, Shop API orqali `DELETE /payment/reversal/{service_id}/{payment_id}`. Lekin Click kabinetida bu funksiya yoqilgan bo'lishi kerak.

### S: Click va Payme orasidagi farq?
**J:**
- **Click:** so'm, MD5 signature, action 0/1
- **Payme:** tiyin, Basic Auth, JSON-RPC (CheckPerformTransaction, CreateTransaction, PerformTransaction, CancelTransaction)

---

## Foydali Havolalar

- Click rasmiy hujjat: https://docs.click.uz
- Click kabinet: https://merchant.click.uz
- Test muhit: https://test.click.uz

---

**Yaratilgan:** 2026-05-23

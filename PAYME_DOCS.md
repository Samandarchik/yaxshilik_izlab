# Payme To'lov Tizimi — To'liq Hujjat

Payme — O'zbekistondagi yirik to'lov tizimi (Click bilan birga). Bu hujjat backend integratsiyasi uchun mo'ljallangan.

> ⚠️ **DIQQAT:** Payme **tiyin** da ishlaydi. 1 so'm = 100 tiyin.
> Ya'ni 100 000 so'm = `10000000` tiyin.

---

## Mundarija

1. [Asosiy Tushunchalar](#asosiy-tushunchalar)
2. [Hisob Ma'lumotlari (Credentials)](#hisob-malumotlari)
3. [Autentifikatsiya (Basic Auth)](#autentifikatsiya)
4. [Merchant API — Webhook Metodlari](#merchant-api)
5. [Subscribe API (Saqlangan Kartalar)](#subscribe-api)
6. [Standart Oqim (Flow)](#standart-oqim)
7. [Xatolik Kodlari](#xatolik-kodlari)
8. [Tranzaksiya Holatlari](#tranzaksiya-holatlari)
9. [Test Kartalar](#test-kartalar)
10. [Xavfsizlik](#xavfsizlik)
11. [Tez-tez Uchraydigan Savollar](#faq)

---

## Asosiy Tushunchalar

| Termin | Ma'nosi |
|--------|---------|
| `merchant_id` | Payme tomonidan beriladigan kassa ID si |
| `key` | Test va Production uchun alohida maxfiy kalit |
| `id` | Payme tranzaksiya ID si (24 ta belgi, MongoDB ObjectId) |
| `account` | Tranzaksiyani tegishli ob'ektga bog'lash (masalan `{order_id: 123}`) |
| `amount` | To'lov summasi (**tiyin** da, masalan: `10000000` = 100 000 so'm) |
| `time` | Unix timestamp **millisekund** da (masalan `1716451200000`) |
| `state` | Tranzaksiya holati (1, 2, -1, -2) |

### Protokol: JSON-RPC 2.0

Payme **JSON-RPC** protokolini ishlatadi (Click dan farqli). Har bir so'rov shunday ko'rinishda:

```json
{
  "jsonrpc": "2.0",
  "id": 12345,
  "method": "MethodName",
  "params": { ... }
}
```

Javob esa:

```json
{
  "jsonrpc": "2.0",
  "id": 12345,
  "result": { ... }
}
```

Yoki xatolik bo'lsa:

```json
{
  "jsonrpc": "2.0",
  "id": 12345,
  "error": {
    "code": -31050,
    "message": {
      "uz": "Buyurtma topilmadi",
      "ru": "Заказ не найден",
      "en": "Order not found"
    }
  }
}
```

---

## Hisob Ma'lumotlari

Payme bilan shartnoma tuzgandan keyin quyidagilar beriladi:

```
MERCHANT_ID  = 5e730e8e0b852a417aa49ceb   (24 belgi)
TEST_KEY     = test_key_here              (sandbox uchun)
PROD_KEY     = production_key_here        (real to'lovlar uchun)
```

`.env` faylida:

```bash
PAYME_MERCHANT_ID=5e730e8e0b852a417aa49ceb
PAYME_KEY=your_secret_key_here
PAYME_ENDPOINT=https://checkout.paycom.uz/api
```

---

## Autentifikatsiya

Payme **Basic Auth** ishlatadi (Click ning MD5 imzosidan farqli).

### Webhook (Merchant API) uchun

Payme sizning serveringizga so'rov yuborganda, `Authorization` headerida quyidagicha keladi:

```
Authorization: Basic Base64(Paycom:KEY)
```

Bu yerda `Paycom` — qattiq belgilangan login, `KEY` — sizning maxfiy kalitingiz.

### Imzo tekshirish:

```python
import base64
from fastapi import Header, HTTPException

PAYME_KEY = "your_secret_key"

def verify_payme_auth(authorization: str = Header(None)):
    if not authorization or not authorization.startswith("Basic "):
        raise HTTPException(401, "Unauthorized")

    encoded = authorization.replace("Basic ", "")
    try:
        decoded = base64.b64decode(encoded).decode()
        login, key = decoded.split(":", 1)
    except Exception:
        raise HTTPException(401, "Invalid auth format")

    if login != "Paycom" or key != PAYME_KEY:
        raise HTTPException(401, "Wrong credentials")
```

---

## Merchant API

Payme webhook ga **bitta** endpoint kerak. Misol uchun:

```
https://yoursite.uz/api/payme
```

Bu endpoint ga Payme **7 ta turli metod** yuborishi mumkin:

1. `CheckPerformTransaction` — to'lov mumkinligini tekshirish
2. `CreateTransaction` — tranzaksiya yaratish
3. `PerformTransaction` — to'lovni amalga oshirish
4. `CancelTransaction` — to'lovni bekor qilish
5. `CheckTransaction` — tranzaksiya holatini tekshirish
6. `GetStatement` — hisobot olish
7. `ChangePassword` — parolni o'zgartirish (kamdan-kam)

---

### 1. CheckPerformTransaction

**Maqsad:** To'lov mumkinligini tekshirish (user mavjudmi, summa to'g'rimi).

**Kiruvchi so'rov:**

```bash
curl -X POST 'https://yoursite.uz/api/payme' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Basic UGF5Y29tOnlvdXJfc2VjcmV0X2tleQ==' \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "CheckPerformTransaction",
    "params": {
      "amount": 10000000,
      "account": {
        "order_id": "123"
      }
    }
  }'
```

**Muvaffaqiyatli javob:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "allow": true
  }
}
```

**Xatolik javobi (buyurtma topilmadi):**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -31050,
    "message": {
      "uz": "Buyurtma topilmadi",
      "ru": "Заказ не найден",
      "en": "Order not found"
    }
  }
}
```

---

### 2. CreateTransaction

**Maqsad:** Tranzaksiyani yaratish (DB ga yozish).

**Kiruvchi so'rov:**

```bash
curl -X POST 'https://yoursite.uz/api/payme' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Basic UGF5Y29tOnlvdXJfc2VjcmV0X2tleQ==' \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "CreateTransaction",
    "params": {
      "id": "5e730e8e0b852a417aa49ceb",
      "time": 1716451200000,
      "amount": 10000000,
      "account": {
        "order_id": "123"
      }
    }
  }'
```

**Javob:**

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "create_time": 1716451200000,
    "transaction": "123",
    "state": 1
  }
}
```

| Maydon | Ma'nosi |
|--------|---------|
| `create_time` | Sizning DB dagi yaratilish vaqti (ms) |
| `transaction` | **SIZNING** DB dagi tranzaksiya ID (string) |
| `state` | `1` = yaratildi, kutilmoqda |

---

### 3. PerformTransaction

**Maqsad:** To'lovni yakunlash (balance qo'shish, buyurtmani faollashtirish).

**Kiruvchi so'rov:**

```bash
curl -X POST 'https://yoursite.uz/api/payme' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Basic UGF5Y29tOnlvdXJfc2VjcmV0X2tleQ==' \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "PerformTransaction",
    "params": {
      "id": "5e730e8e0b852a417aa49ceb"
    }
  }'
```

**Javob:**

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "perform_time": 1716451260000,
    "transaction": "123",
    "state": 2
  }
}
```

`state: 2` — to'lov muvaffaqiyatli amalga oshirildi.

---

### 4. CancelTransaction

**Maqsad:** Tranzaksiyani bekor qilish (refund yoki pending dan voz kechish).

**Kiruvchi so'rov:**

```bash
curl -X POST 'https://yoursite.uz/api/payme' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Basic UGF5Y29tOnlvdXJfc2VjcmV0X2tleQ==' \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "CancelTransaction",
    "params": {
      "id": "5e730e8e0b852a417aa49ceb",
      "reason": 5
    }
  }'
```

| `reason` | Ma'nosi |
|----------|---------|
| `1` | Foydalanuvchi tomonidan bekor qilindi |
| `2` | Tranzaksiya yakunlanmadi |
| `3` | Texnik xatolik |
| `4` | Buyurtma to'lab bo'lingan (refund) |
| `5` | Boshqa sababga ko'ra |

**Javob:**

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "cancel_time": 1716451300000,
    "transaction": "123",
    "state": -1
  }
}
```

| `state` | Ma'nosi |
|---------|---------|
| `-1` | Yaratilgan, lekin bekor qilingan |
| `-2` | Yakunlangan, keyin bekor qilingan (refund) |

---

### 5. CheckTransaction

**Maqsad:** Tranzaksiya holatini bilish (Payme har gal so'raydi).

**Kiruvchi so'rov:**

```bash
curl -X POST 'https://yoursite.uz/api/payme' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Basic UGF5Y29tOnlvdXJfc2VjcmV0X2tleQ==' \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "CheckTransaction",
    "params": {
      "id": "5e730e8e0b852a417aa49ceb"
    }
  }'
```

**Javob:**

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "create_time": 1716451200000,
    "perform_time": 1716451260000,
    "cancel_time": 0,
    "transaction": "123",
    "state": 2,
    "reason": null
  }
}
```

---

### 6. GetStatement

**Maqsad:** Vaqt oralig'ida bo'lgan tranzaksiyalarni olish.

```bash
curl -X POST 'https://yoursite.uz/api/payme' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Basic UGF5Y29tOnlvdXJfc2VjcmV0X2tleQ==' \
  -d '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "GetStatement",
    "params": {
      "from": 1716000000000,
      "to":   1716500000000
    }
  }'
```

**Javob:**

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "result": {
    "transactions": [
      {
        "id": "5e730e8e0b852a417aa49ceb",
        "time": 1716451200000,
        "amount": 10000000,
        "account": { "order_id": "123" },
        "create_time": 1716451200000,
        "perform_time": 1716451260000,
        "cancel_time": 0,
        "transaction": "123",
        "state": 2,
        "reason": null
      }
    ]
  }
}
```

---

## Subscribe API

Subscribe API — bu **server tomondan** to'g'ridan-to'g'ri to'lov qilish (saqlangan karta orqali, "Receipt" yaratish va boshqalar). Bu Merchant API dan **farqli** — bu yerda **siz Payme ga** so'rov yuborasiz.

**Base URL:**
- Production: `https://checkout.paycom.uz/api`
- Test: `https://checkout.test.paycom.uz/api`

**Auth header:**
```
X-Auth: MERCHANT_ID:KEY
```

(yoki ba'zi metodlarda `MERCHANT_ID:USER_KEY`)

---

### 1. Receipt yaratish (`receipts.create`)

```bash
curl -X POST 'https://checkout.paycom.uz/api' \
  -H 'Content-Type: application/json' \
  -H 'X-Auth: 5e730e8e0b852a417aa49ceb:your_secret_key' \
  -d '{
    "id": 1,
    "method": "receipts.create",
    "params": {
      "amount": 10000000,
      "account": {
        "order_id": "123"
      }
    }
  }'
```

**Javob:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "receipt": {
      "_id": "5e730e8e0b852a417aa49ceb",
      "create_time": 1716451200000,
      "amount": 10000000,
      "account": [
        { "name": "order_id", "title": "Buyurtma ID", "value": "123" }
      ],
      "state": 0,
      "pay_time": 0,
      "cancel_time": 0
    }
  }
}
```

| `state` | Ma'nosi |
|---------|---------|
| `0` | Yaratilgan |
| `1` | Tasdiqlangan kutilmoqda |
| `2` | Tasdiqlangan |
| `3` | Bekor qilingan |
| `4` | To'langan |
| `50` | Pulni qaytarish kerak |

---

### 2. Karta yaratish (`cards.create`)

```bash
curl -X POST 'https://checkout.paycom.uz/api' \
  -H 'Content-Type: application/json' \
  -H 'X-Auth: 5e730e8e0b852a417aa49ceb:your_secret_key' \
  -d '{
    "id": 1,
    "method": "cards.create",
    "params": {
      "card": {
        "number": "8600495473316478",
        "expire": "0399"
      },
      "save": true
    }
  }'
```

**Javob:**

```json
{
  "result": {
    "card": {
      "number": "860049******6478",
      "expire": "0399",
      "token": "627a8088e28e74797e9aab7d_DjzMrJZ6oRzs...",
      "recurrent": true,
      "verify": false
    }
  }
}
```

`token` — keyinchalik shu karta bilan to'lov qilishda ishlatiladi.

---

### 3. Karta tasdig'i uchun SMS yuborish (`cards.get_verify_code`)

```bash
curl -X POST 'https://checkout.paycom.uz/api' \
  -H 'Content-Type: application/json' \
  -H 'X-Auth: 5e730e8e0b852a417aa49ceb:your_secret_key' \
  -d '{
    "id": 1,
    "method": "cards.get_verify_code",
    "params": {
      "token": "627a8088e28e74797e9aab7d_DjzMrJZ6oRzs..."
    }
  }'
```

---

### 4. SMS kodni tekshirish (`cards.verify`)

```bash
curl -X POST 'https://checkout.paycom.uz/api' \
  -H 'Content-Type: application/json' \
  -H 'X-Auth: 5e730e8e0b852a417aa49ceb:your_secret_key' \
  -d '{
    "id": 1,
    "method": "cards.verify",
    "params": {
      "token": "627a8088e28e74797e9aab7d_DjzMrJZ6oRzs...",
      "code": "666666"
    }
  }'
```

---

### 5. Receipt ni to'lash (`receipts.pay`)

```bash
curl -X POST 'https://checkout.paycom.uz/api' \
  -H 'Content-Type: application/json' \
  -H 'X-Auth: 5e730e8e0b852a417aa49ceb:your_secret_key' \
  -d '{
    "id": 1,
    "method": "receipts.pay",
    "params": {
      "id": "5e730e8e0b852a417aa49ceb",
      "token": "627a8088e28e74797e9aab7d_DjzMrJZ6oRzs..."
    }
  }'
```

---

### 6. Boshqa foydali metodlar

| Metod | Maqsad |
|-------|--------|
| `receipts.send` | To'lov chekini SMS/Email orqali yuborish |
| `receipts.cancel` | Receipt ni bekor qilish |
| `receipts.check` | Receipt holatini tekshirish |
| `receipts.get` | Receipt ma'lumotlarini olish |
| `receipts.get_all` | Receipt ro'yxati |
| `cards.check` | Karta holatini tekshirish |
| `cards.remove` | Saqlangan kartani o'chirish |

---

## Standart Oqim

### Variant 1: Checkout sahifasi (eng oddiy)

```
┌──────────┐                ┌──────────┐               ┌──────────┐
│  USER    │                │ BACKEND  │               │  PAYME   │
└────┬─────┘                └────┬─────┘               └────┬─────┘
     │                           │                          │
     │ 1. "100 000 to'ldirish"   │                          │
     ├──────────────────────────>│                          │
     │                           │                          │
     │                           │ 2. DB ga yozish:         │
     │                           │   order_id=123           │
     │                           │   amount=10000000 (tiyin)│
     │                           │                          │
     │                           │ 3. Checkout URL yaratish │
     │ 4. Redirect Payme ga      │                          │
     │<──────────────────────────┤                          │
     │                                                      │
     │ 5. Karta ma'lumotlari + SMS                          │
     ├─────────────────────────────────────────────────────>│
     │                                                      │
     │                           │ 6. CheckPerformTransaction│
     │                           │<─────────────────────────┤
     │                           │   allow=true             │
     │                           ├─────────────────────────>│
     │                           │                          │
     │                           │ 7. CreateTransaction     │
     │                           │<─────────────────────────┤
     │                           │   state=1                │
     │                           ├─────────────────────────>│
     │                           │                          │
     │                           │ 8. PerformTransaction    │
     │                           │<─────────────────────────┤
     │                           │   state=2                │
     │                           │   balance += 100K        │
     │                           ├─────────────────────────>│
     │                                                      │
     │ 9. Success page                                      │
     │<─────────────────────────────────────────────────────┤
```

---

## Checkout URL Yaratish

Userni Payme to'lov sahifasiga shu URL orqali yuborasiz:

```
https://checkout.paycom.uz/{BASE64_PARAMS}
```

`BASE64_PARAMS` — quyidagi qatorni Base64 ga aylantirish:

```
m=MERCHANT_ID;ac.order_id=123;a=10000000;c=https://yoursite.uz/success
```

### Python misol:

```python
import base64

def make_checkout_url(merchant_id: str, order_id: str, amount: int, return_url: str) -> str:
    # amount — TIYIN da!
    params = f"m={merchant_id};ac.order_id={order_id};a={amount};c={return_url}"
    encoded = base64.b64encode(params.encode()).decode()
    return f"https://checkout.paycom.uz/{encoded}"

url = make_checkout_url(
    merchant_id="5e730e8e0b852a417aa49ceb",
    order_id="123",
    amount=10000000,         # 100 000 so'm
    return_url="https://yoursite.uz/success"
)
# https://checkout.paycom.uz/bT01ZTczMGU4ZTBiODUyYTQxN2FhNDljZWI7...
```

### Yoki POST forma orqali:

```html
<form method="POST" action="https://checkout.paycom.uz/">
  <input type="hidden" name="merchant" value="5e730e8e0b852a417aa49ceb">
  <input type="hidden" name="amount" value="10000000">
  <input type="hidden" name="account[order_id]" value="123">
  <input type="hidden" name="callback" value="https://yoursite.uz/success">
  <input type="hidden" name="callback_timeout" value="15000">
  <input type="hidden" name="lang" value="uz">
  <button type="submit">Payme orqali to'lash</button>
</form>
```

---

## To'liq Python (FastAPI) Misol

```python
import base64
from datetime import datetime
from fastapi import FastAPI, Request, Header, HTTPException
from sqlalchemy.orm import Session

app = FastAPI()

PAYME_MERCHANT_ID = "5e730e8e0b852a417aa49ceb"
PAYME_KEY = "your_secret_key"

# --- Xatolik kodlari ---
class PaymeError:
    INVALID_AMOUNT       = -31001
    ORDER_NOT_FOUND      = -31050
    CANNOT_PERFORM       = -31008
    ORDER_ALREADY_PAID   = -31051
    TRANSACTION_NOT_FOUND= -31003
    INVALID_AUTH         = -32504
    METHOD_NOT_FOUND     = -32601


def error(code: int, message: str, data: str = None):
    err = {
        "code": code,
        "message": {"uz": message, "ru": message, "en": message},
    }
    if data:
        err["data"] = data
    return err


# --- Auth tekshirish ---
def check_auth(authorization: str):
    if not authorization or not authorization.startswith("Basic "):
        return False
    try:
        decoded = base64.b64decode(authorization[6:]).decode()
        login, key = decoded.split(":", 1)
        return login == "Paycom" and key == PAYME_KEY
    except Exception:
        return False


# --- Asosiy webhook ---
@app.post("/api/payme")
async def payme_webhook(
    request: Request,
    authorization: str = Header(None),
    db: Session = ...,
):
    body = await request.json()
    rpc_id = body.get("id")
    method = body.get("method")
    params = body.get("params", {})

    # 1. Auth
    if not check_auth(authorization):
        return {
            "jsonrpc": "2.0",
            "id": rpc_id,
            "error": error(PaymeError.INVALID_AUTH, "Insufficient privileges"),
        }

    # 2. Metodga qarab yo'naltirish
    if method == "CheckPerformTransaction":
        return check_perform(rpc_id, params, db)
    elif method == "CreateTransaction":
        return create_transaction(rpc_id, params, db)
    elif method == "PerformTransaction":
        return perform_transaction(rpc_id, params, db)
    elif method == "CancelTransaction":
        return cancel_transaction(rpc_id, params, db)
    elif method == "CheckTransaction":
        return check_transaction(rpc_id, params, db)
    elif method == "GetStatement":
        return get_statement(rpc_id, params, db)
    else:
        return {
            "jsonrpc": "2.0",
            "id": rpc_id,
            "error": error(PaymeError.METHOD_NOT_FOUND, "Method not found"),
        }


# --- 1. CheckPerformTransaction ---
def check_perform(rpc_id, params, db):
    order_id = params.get("account", {}).get("order_id")
    amount = params.get("amount")

    order = db.query(Order).filter_by(id=order_id).first()

    if not order:
        return {"jsonrpc": "2.0", "id": rpc_id,
                "error": error(PaymeError.ORDER_NOT_FOUND, "Order not found")}

    if order.amount * 100 != amount:  # so'm → tiyin
        return {"jsonrpc": "2.0", "id": rpc_id,
                "error": error(PaymeError.INVALID_AMOUNT, "Incorrect amount")}

    if order.status == "paid":
        return {"jsonrpc": "2.0", "id": rpc_id,
                "error": error(PaymeError.ORDER_ALREADY_PAID, "Already paid")}

    return {"jsonrpc": "2.0", "id": rpc_id, "result": {"allow": True}}


# --- 2. CreateTransaction ---
def create_transaction(rpc_id, params, db):
    payme_id = params["id"]
    time_ms = params["time"]
    amount = params["amount"]
    order_id = params.get("account", {}).get("order_id")

    # Allaqachon mavjudligini tekshirish
    tx = db.query(Transaction).filter_by(payme_id=payme_id).first()

    if tx:
        if tx.state != 1:
            return {"jsonrpc": "2.0", "id": rpc_id,
                    "error": error(PaymeError.CANNOT_PERFORM, "Cannot perform")}
        return {
            "jsonrpc": "2.0", "id": rpc_id,
            "result": {
                "create_time": tx.create_time,
                "transaction": str(tx.id),
                "state": tx.state,
            },
        }

    # CheckPerform mantiqi
    order = db.query(Order).filter_by(id=order_id).first()
    if not order or order.amount * 100 != amount:
        return {"jsonrpc": "2.0", "id": rpc_id,
                "error": error(PaymeError.ORDER_NOT_FOUND, "Order not found")}

    # Yangi tranzaksiya
    tx = Transaction(
        payme_id=payme_id,
        order_id=order_id,
        amount=amount,
        state=1,
        create_time=time_ms,
    )
    db.add(tx)
    db.commit()

    return {
        "jsonrpc": "2.0", "id": rpc_id,
        "result": {
            "create_time": time_ms,
            "transaction": str(tx.id),
            "state": 1,
        },
    }


# --- 3. PerformTransaction ---
def perform_transaction(rpc_id, params, db):
    tx = db.query(Transaction).filter_by(payme_id=params["id"]).first()

    if not tx:
        return {"jsonrpc": "2.0", "id": rpc_id,
                "error": error(PaymeError.TRANSACTION_NOT_FOUND, "Transaction not found")}

    if tx.state == 2:
        # Allaqachon to'langan — idempotency
        return {
            "jsonrpc": "2.0", "id": rpc_id,
            "result": {
                "perform_time": tx.perform_time,
                "transaction": str(tx.id),
                "state": 2,
            },
        }

    if tx.state != 1:
        return {"jsonrpc": "2.0", "id": rpc_id,
                "error": error(PaymeError.CANNOT_PERFORM, "Cannot perform")}

    now_ms = int(datetime.now().timestamp() * 1000)
    tx.state = 2
    tx.perform_time = now_ms

    # Buyurtmani faollashtirish / balance qo'shish
    order = db.query(Order).filter_by(id=tx.order_id).first()
    order.status = "paid"
    order.user.balance += tx.amount // 100  # tiyin → so'm

    db.commit()

    return {
        "jsonrpc": "2.0", "id": rpc_id,
        "result": {
            "perform_time": now_ms,
            "transaction": str(tx.id),
            "state": 2,
        },
    }


# --- 4. CancelTransaction ---
def cancel_transaction(rpc_id, params, db):
    tx = db.query(Transaction).filter_by(payme_id=params["id"]).first()
    if not tx:
        return {"jsonrpc": "2.0", "id": rpc_id,
                "error": error(PaymeError.TRANSACTION_NOT_FOUND, "Transaction not found")}

    now_ms = int(datetime.now().timestamp() * 1000)

    if tx.state == 1:
        tx.state = -1
        tx.cancel_time = now_ms
    elif tx.state == 2:
        # Refund logikasi (agar mumkin bo'lsa)
        tx.state = -2
        tx.cancel_time = now_ms
        # tx.user.balance -= tx.amount // 100

    tx.reason = params.get("reason")
    db.commit()

    return {
        "jsonrpc": "2.0", "id": rpc_id,
        "result": {
            "cancel_time": tx.cancel_time,
            "transaction": str(tx.id),
            "state": tx.state,
        },
    }


# --- 5. CheckTransaction ---
def check_transaction(rpc_id, params, db):
    tx = db.query(Transaction).filter_by(payme_id=params["id"]).first()
    if not tx:
        return {"jsonrpc": "2.0", "id": rpc_id,
                "error": error(PaymeError.TRANSACTION_NOT_FOUND, "Transaction not found")}

    return {
        "jsonrpc": "2.0", "id": rpc_id,
        "result": {
            "create_time": tx.create_time or 0,
            "perform_time": tx.perform_time or 0,
            "cancel_time": tx.cancel_time or 0,
            "transaction": str(tx.id),
            "state": tx.state,
            "reason": tx.reason,
        },
    }


# --- 6. GetStatement ---
def get_statement(rpc_id, params, db):
    from_ts = params["from"]
    to_ts = params["to"]

    txs = db.query(Transaction).filter(
        Transaction.create_time >= from_ts,
        Transaction.create_time <= to_ts,
    ).all()

    return {
        "jsonrpc": "2.0", "id": rpc_id,
        "result": {
            "transactions": [
                {
                    "id": tx.payme_id,
                    "time": tx.create_time,
                    "amount": tx.amount,
                    "account": {"order_id": tx.order_id},
                    "create_time": tx.create_time or 0,
                    "perform_time": tx.perform_time or 0,
                    "cancel_time": tx.cancel_time or 0,
                    "transaction": str(tx.id),
                    "state": tx.state,
                    "reason": tx.reason,
                }
                for tx in txs
            ]
        },
    }


# --- 7. To'lov yaratish (foydalanuvchi tomonidan chaqiriladi) ---
@app.post("/api/payment/create")
async def create_payment(amount: int, user_id: int, db: Session = ...):
    # amount — so'm da
    if amount < 1000 or amount > 10_000_000:
        raise HTTPException(400, "Summa 1 000 dan 10 000 000 gacha bo'lishi kerak")

    order = Order(user_id=user_id, amount=amount, status="pending")
    db.add(order)
    db.commit()

    # Checkout URL
    params_str = (
        f"m={PAYME_MERCHANT_ID};"
        f"ac.order_id={order.id};"
        f"a={amount * 100};"   # tiyin ga aylantirish
        f"c=https://yoursite.uz/success"
    )
    encoded = base64.b64encode(params_str.encode()).decode()

    return {"payment_url": f"https://checkout.paycom.uz/{encoded}"}
```

---

## Xatolik Kodlari

### Standart JSON-RPC xatoliklari

| Kod | Ma'nosi |
|-----|---------|
| `-32700` | Parse error |
| `-32600` | Invalid request |
| `-32601` | Method not found |
| `-32602` | Invalid params |
| `-32603` | Internal error |
| `-32504` | Insufficient privileges / Authorization xatosi |

### Payme business xatoliklari

| Kod | Ma'nosi | Qachon |
|-----|---------|--------|
| `-31001` | Noto'g'ri summa | Amount mos kelmadi |
| `-31003` | Tranzaksiya topilmadi | `id` bo'yicha topilmadi |
| `-31007` | Buyurtmani bekor qilib bo'lmaydi | Holatga ko'ra |
| `-31008` | Operatsiyani bajarib bo'lmaydi | State noto'g'ri |
| `-31050` | Buyurtma topilmadi | `account.order_id` noto'g'ri |
| `-31051` | Buyurtma allaqachon to'langan | Status `paid` |
| `-31052` | Buyurtma to'lab bo'lmas holatda | Bekor qilingan, eskirgan |
| `-31053` | Buyurtma yopilgan | |
| `-31054` | Operatsiya muddati o'tgan | Timeout |
| `-31055` | Internal server error | Sizning xatolik |
| `-31099` | Foydalanuvchi xatosi | |

---

## Tranzaksiya Holatlari

| `state` | Ma'nosi |
|---------|---------|
| `1` | Yaratilgan, to'lov kutilmoqda |
| `2` | Muvaffaqiyatli to'langan |
| `-1` | Yaratilgan, lekin bekor qilingan |
| `-2` | To'langan, keyin refund qilingan |

State transitions:

```
       CreateTransaction
[NULL] ─────────────────> [1]
                           │
                           ├── PerformTransaction ──> [2]
                           │                           │
                           │                           ├── CancelTransaction ──> [-2]
                           │
                           └── CancelTransaction ───> [-1]
```

---

## Test Kartalar

Sandbox muhitida ishlatish uchun:

```
Karta raqami:   8600 4954 7331 6478
Amal qilish:    03/99
SMS kod:        666666
```

Sandbox URL:
```
Production: https://checkout.paycom.uz
Sandbox:    https://checkout.test.paycom.uz
```

---

## Xavfsizlik

### 1. HTTPS shart
Payme webhook faqat HTTPS bilan ishlaydi.

### 2. Basic Auth ni MAJBURIY tekshiring
```python
if not check_auth(authorization):
    return error(-32504, "Insufficient privileges")
```

### 3. Payme yuborgan `amount` ga ishonmang
Har doim DB dagi `order.amount * 100` bilan solishtiring (so'm → tiyin).

### 4. Idempotency
Payme bir tranzaksiyani **bir necha marta** yuborishi mumkin (network retry). Sizning kod:
- Mavjud tranzaksiyani topishi
- Holatiga qarab to'g'ri javob qaytarishi kerak

```python
if tx.state == 2:
    # Allaqachon to'langan — xato emas, success qaytarish
    return success_with_existing_data
```

### 5. State machine ni qattiq saqlang
```
1 → 2  (perform)
1 → -1 (cancel)
2 → -2 (refund)
```
Boshqa o'tishlar XATO.

### 6. Timeout va race condition
`CreateTransaction` da `time` ni tekshiring — 12 soatdan eski bo'lsa, rad eting.

```python
from datetime import datetime
now_ms = int(datetime.now().timestamp() * 1000)
if now_ms - params["time"] > 12 * 60 * 60 * 1000:
    return error(-31008, "Transaction expired")
```

### 7. IP whitelist (qo'shimcha himoya)
Payme serverlari (so'rab oling) faqat shu IP lardan keladi.

---

## DB Schema Tavsiya

```sql
CREATE TABLE orders (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    amount      DECIMAL(15, 2) NOT NULL,    -- so'm da
    status      VARCHAR(20) DEFAULT 'pending',
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE transactions (
    id              BIGSERIAL PRIMARY KEY,
    payme_id        VARCHAR(50) UNIQUE NOT NULL,  -- Payme ning 24 belgili ID si
    order_id        BIGINT REFERENCES orders(id),
    amount          BIGINT NOT NULL,              -- tiyin da
    state           SMALLINT NOT NULL,            -- 1, 2, -1, -2
    create_time     BIGINT,                       -- ms
    perform_time    BIGINT,
    cancel_time     BIGINT,
    reason          SMALLINT,

    INDEX idx_payme_id (payme_id),
    INDEX idx_order_id (order_id),
    INDEX idx_create_time (create_time)
);
```

---

## FAQ

### S: Payme va Click orasidagi farq nima?

| | Click | Payme |
|---|---|---|
| Birlik | so'm | **tiyin** |
| Protokol | Form-encoded | **JSON-RPC 2.0** |
| Auth | MD5 signature | **Basic Auth** |
| Webhook so'rovlari | 2 ta (Prepare, Complete) | **5-7 ta metod** |
| Tranzaksiya ID | merchant_trans_id | account.order_id + Payme id |
| Vaqt formati | `YYYY-MM-DD HH:MM:SS` | Unix timestamp **ms** |

### S: User istalgan summani kirita oladimi?
**J:** Ha. Frontend → Backend (DB ga yozadi) → Payme. Min/Max limit qo'ying.

### S: 100 000 so'm Payme da qancha bo'ladi?
**J:** `10000000` tiyin (100 000 × 100).

### S: Nima uchun Payme bir tranzaksiyani 5 marta yuborishi mumkin?
**J:** Network reliability uchun. Sizning kod idempotent bo'lishi kerak (`state` ni tekshirib, to'g'ri javob qaytarish).

### S: `CheckPerformTransaction` qachon chaqiriladi?
**J:** User Payme da karta ma'lumotlarini kiritgandan keyin, SMS kod yuborishdan **oldin**. Shu yerda buyurtma mavjudligini tekshirasiz.

### S: Userning to'lov sahifasi qanday ko'rinadi?
**J:** Payme o'zining oq-yashil checkout sahifasini ko'rsatadi. Siz unga **redirect** qilasiz, bo'lgani.

### S: Refund qilish mumkinmi?
**J:** Ha, `CancelTransaction` ni `state=2` bo'lganda chaqirsangiz, state `-2` ga o'tadi. Lekin Payme kabinetida bu funksiya yoqilgan bo'lishi kerak.

### S: Subscribe API qachon kerak?
**J:** Agar siz:
- Kartalarni saqlamoqchi bo'lsangiz (recurring payments)
- O'z saytingizda checkout qilmoqchi bo'lsangiz (Payme sahifasiga redirect qilmasdan)
- Avtomatik to'lovlar (obuna) qilsangiz

### S: Test va Production kalitlari boshqa-boshqami?
**J:** Ha. Test (`https://checkout.test.paycom.uz`) va Production (`https://checkout.paycom.uz`) uchun **alohida key** beriladi.

---

## Foydali Havolalar

- Payme rasmiy hujjat: https://developer.help.paycom.uz
- Payme kabinet: https://merchant.paycom.uz
- Test muhit: https://checkout.test.paycom.uz
- Subscribe API: https://developer.help.paycom.uz/metody-subscribe-api

---

**Yaratilgan:** 2026-05-23

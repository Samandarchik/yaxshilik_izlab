-- Admin users
CREATE TABLE IF NOT EXISTS admins (
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name     VARCHAR(255) DEFAULT '',
    role          VARCHAR(20) DEFAULT 'admin',
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- People (bemorlar / hikoyalar)
CREATE TABLE IF NOT EXISTS people (
    id                BIGSERIAL PRIMARY KEY,
    name              VARCHAR(255) NOT NULL,
    age               INT DEFAULT 0,
    region            VARCHAR(120) DEFAULT '',
    diagnosis         VARCHAR(255) DEFAULT '',
    help              VARCHAR(255) DEFAULT '',
    facility          VARCHAR(255) DEFAULT '',
    facility_verified BOOLEAN DEFAULT FALSE,
    org               VARCHAR(255) DEFAULT '',
    story             TEXT DEFAULT '',
    photo_url         TEXT DEFAULT '',
    target            BIGINT NOT NULL,           -- so'm
    raised            BIGINT NOT NULL DEFAULT 0, -- so'm
    donors            INT NOT NULL DEFAULT 0,
    days_left         INT DEFAULT 30,
    urgent            BOOLEAN DEFAULT FALSE,
    category          VARCHAR(80) DEFAULT '',
    author_name       VARCHAR(255) DEFAULT '',
    author_role       VARCHAR(255) DEFAULT '',
    status            VARCHAR(20) DEFAULT 'active', -- active | paused | closed
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_people_status ON people(status);
CREATE INDEX IF NOT EXISTS idx_people_urgent ON people(urgent);

-- Donations / payments (Click + Payme bir jadvalda)
CREATE TABLE IF NOT EXISTS donations (
    id            BIGSERIAL PRIMARY KEY,
    person_id     BIGINT NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    provider      VARCHAR(10) NOT NULL,          -- 'click' | 'payme'
    amount_tiyin  BIGINT NOT NULL,               -- ichki normalize: tiyin
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
                                                  -- pending | prepared | paid | cancelled | failed
    anonim        BOOLEAN DEFAULT FALSE,
    donor_name    VARCHAR(255) DEFAULT '',
    donor_phone   VARCHAR(32)  DEFAULT '',

    -- Click polya
    click_trans_id      VARCHAR(64),
    click_paydoc_id     VARCHAR(64),
    merchant_prepare_id BIGINT,

    -- Payme polya
    payme_id            VARCHAR(64) UNIQUE,
    payme_state         SMALLINT,           -- 1, 2, -1, -2
    payme_create_time   BIGINT,             -- ms
    payme_perform_time  BIGINT,
    payme_cancel_time   BIGINT,
    payme_reason        SMALLINT,

    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    paid_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_donations_person     ON donations(person_id);
CREATE INDEX IF NOT EXISTS idx_donations_status     ON donations(status);
CREATE INDEX IF NOT EXISTS idx_donations_provider   ON donations(provider);
CREATE INDEX IF NOT EXISTS idx_donations_created_at ON donations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_donations_payme_id   ON donations(payme_id);
CREATE INDEX IF NOT EXISTS idx_donations_click_tid  ON donations(click_trans_id);

-- Telegram Mini App: yordam bergan foydalanuvchini aniqlash uchun
ALTER TABLE donations ADD COLUMN IF NOT EXISTS tg_user_id  BIGINT;
ALTER TABLE donations ADD COLUMN IF NOT EXISTS tg_username VARCHAR(64);
CREATE INDEX IF NOT EXISTS idx_donations_tg_user ON donations(tg_user_id);

-- Audit log (optional, foydali)
CREATE TABLE IF NOT EXISTS audit_log (
    id         BIGSERIAL PRIMARY KEY,
    actor      VARCHAR(255) DEFAULT '',
    action     VARCHAR(120) NOT NULL,
    target     VARCHAR(120) DEFAULT '',
    payload    JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

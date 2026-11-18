CREATE TABLE IF NOT EXISTS roles (
    name        VARCHAR(20)  PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
    ('admin',        'Kelola akun dokter, jadwal & kuota, laporan, monitoring sistem'),
    ('doctor',       'Kelola antrean pasiennya, tulis rekam medis, balas ulasan'),
    ('receptionist', 'Verifikasi kedatangan, update status antrean, rekap harian'),
    ('patient',      'Booking & batalkan antrean, baca RME miliknya, buat ulasan')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL    PRIMARY KEY,
    username   VARCHAR(50)  NOT NULL UNIQUE,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    role       VARCHAR(20)  NOT NULL REFERENCES roles(name) ON UPDATE CASCADE,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

-- Refresh token disimpan sebagai HASH SHA-256, bukan nilai aslinya:
-- bila table ini bocor, penyerang tidak mendapat token yang bisa dipakai.
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens (user_id);

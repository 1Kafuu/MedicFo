# MedicFo — Sistem Manajemen & Rekam Medis Rumah Sakit (E-Healthcare Backend)

**MedicFo Backend API** (*Hospital Management & Electronic Health Record System*) adalah sistem backend layanan kesehatan berbasis **Go (Fiber)** dengan prinsip **Clean Architecture**. Sistem menangani dua aspek krusial operasional rumah sakit: transaksi pendaftaran antrean berkonkurensi tinggi (anti-overbooking) dan pengelolaan rekam medis elektronik (RME) yang dinamis.

Polyglot Persistence: **PostgreSQL** untuk data transaksional ACID (akun, jadwal, antrean, RBAC, outbox), **MongoDB** untuk dokumen fleksibel (RME, ulasan dokter), dan **Redis + Asynq** untuk background job notifikasi.

## Fitur Sistem

| ID | Fitur | Role Pengakses |
| --- | --- | --- |
| FT-01 | Autentikasi & Token Rotation (JWT access 15m, refresh 7d, Bcrypt) | Publik, semua role |
| FT-02 | Keamanan Endpoint & Rate Limiting | Publik |
| FT-03 | Manajemen Dokter & Spesialisasi (keyset cursor pagination) | Admin (tulis), semua (baca) |
| FT-04 | Jadwal & Kuota Praktik (optimistic locking) | Admin |
| FT-05 | Booking Antrean (pessimistic locking, anti-overbooking) | Pasien |
| FT-06 | Transactional Outbox Event | Internal |
| FT-07 | Worker & Notifikasi (Asynq, idempoten) | Internal |
| FT-08 | Rekam Medis Elektronik (MongoDB, ownership check) | Dokter (tulis), pasien (baca milik) |
| FT-09 | Ulasan & Rating Dokter (unique index, nested replies) | Pasien, dokter, publik |
| FT-10 | Agregasi Analytics Kesehatan (Mongo pipeline) | Dokter, admin |
| FT-11 | Ekspor Laporan CSV/JSON (content negotiation) | Admin, resepsionis |
| FT-12 | Monitoring Queue & Health Check | Admin |
| FT-13 | Dokumentasi API Swagger | Publik, developer |
| FT-14 | Containerized Deployment (Docker Compose) | DevOps |

Detail mekanisme tiap fitur → bagian "Fitur Sistem" di atas.

## Role & Hak Akses (RBAC)

| Role | Hak Akses Utama |
| --- | --- |
| `admin` | Kelola akun dokter, jadwal & kuota, laporan, monitoring sistem |
| `doctor` | Antrean pasiennya, isi RME, balas ulasan, lihat laporan |
| `receptionist` | Verifikasi kedatangan, update status antrean, ekspor rekap harian |
| `patient` | Booking & batalkan antrean, baca RME miliknya, buat ulasan |

## Rancangan Database

![ERD MedicFo](./DATABASE.ERD.png)

### PostgreSQL

#### `users` — akun & autentikasi

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| username | VARCHAR UNIQUE | index untuk lookup login |
| email | VARCHAR UNIQUE | index untuk lookup login |
| password | VARCHAR | hash Bcrypt cost 12 |
| role | VARCHAR FK → `roles.name` | `admin`/`doctor`/`receptionist`/`patient` |
| is_active | BOOL | soft-disable akun |
| created_at | TIMESTAMPTZ | |

#### `roles` — master role RBAC

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| name | VARCHAR PK | `admin` / `doctor` / `receptionist` / `patient` |
| description | TEXT | |

#### `permissions` — master permission

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| name | VARCHAR PK | contoh: `appointment:create`, `medical_record:write` |
| description | TEXT | |

#### `role_permissions` — pemetaan role ↔ permission

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| role_name | VARCHAR PK, FK → `roles.name` | composite PK `(role_name, permission_name)` |
| permission_name | VARCHAR PK, FK → `permissions.name` | dimuat ke memori sebagai permission set saat autentikasi JWT |

#### `doctors` — profil medis

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| id | SERIAL PK | |
| user_id | INT FK → `users.id` UNIQUE | 1-to-1 dengan users ber-role doctor |
| name | VARCHAR | |
| specialization | VARCHAR | |
| phone | VARCHAR | |

#### `doctor_schedules` — jadwal & kuota

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| id | SERIAL PK | |
| doctor_id | INT FK → `doctors.id` | |
| day_of_week | SMALLINT | 0=Minggu … 6=Sabtu |
| start_time / end_time | TIME | jam praktik |
| quota | INT | kuota harian |
| booked_count | INT | di-increment saat booking, `CHECK (booked_count <= quota)` |
| version | INT | optimistic locking admin (mulai 0, +1 tiap update) |

#### `appointments` — booking antrean

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| patient_id | INT FK → `users.id` | |
| schedule_id | INT FK → `doctor_schedules.id` | |
| appointment_date | DATE | tanggal kunjungan |
| queue_number | INT | dialokasi atomik dalam transaksi |
| status | VARCHAR | `booked` / `checked_in` / `completed` / `cancelled` |
| created_at | TIMESTAMPTZ | |


#### `outbox` — event domain (Transactional Outbox Pattern)

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| event_type | VARCHAR | mis. `appointment.created` |
| payload | JSONB | isi event |
| created_at | TIMESTAMPTZ | |
| published_at | TIMESTAMPTZ NULL | NULL = belum terkirim; relay baca `SKIP LOCKED` → job Asynq → set `published_at` |

#### `notifications` — audit pengiriman worker

| Kolom | Tipe | Catatan |
| --- | --- | --- |
| id | BIGSERIAL PK | |
| outbox_id | BIGINT UNIQUE FK → `outbox.id` | idempotensi consumer: job ganda ditolak DB |
| user_id | FK → `users.id` | penerima |
| channel | VARCHAR | `email` / `sms` |
| subject / body | VARCHAR / TEXT | |
| sent_at | TIMESTAMPTZ | |

### MongoDB

#### `medical_records` — RME pasien

| Field | Tipe |
| --- | --- |
| _id | ObjectId |
| appointment_id / patient_id / doctor_id / doctor_name | Number / String |
| diagnosis | String |
| symptoms | Array of String |
| vital_signs | Object (`blood_pressure`, `temp`, `heart_rate`) |
| prescriptions | Array of Object (`medicine_name`, `dosage`, `notes`) |
| lab_results | Array of Object (`test_name`, `result`, `file_url`) |
| created_at | ISODate |

#### `doctor_reviews` — ulasan & rating

| Field | Tipe |
| --- | --- |
| _id | ObjectId |
| doctor_id / patient_id | Number |
| patient_username | String |
| rating | Number 1–5 |
| comment | String |
| tags | Array of String |
| replies | Array of Object (`_id`, `user_id`, `body`, `created_at`) |
| helpful_count | Number |
| helpful_by | Array of Number |
| created_at | ISODate |

## Anggota & Peran

| Anggota | Peran |
| --- | --- |
| Vicky | Git Master — fondasi, auth & security, merge semua PR |
| Via | Fitur dokter, jadwal & kuota, booking antrean, outbox |
| Cindy | Worker Asynq, monitoring, ekspor laporan, Docker |
| Nadine | RME & ulasan (MongoDB), agregasi, Swagger |


## Cara Menjalankan

```bash
# 1. Salin env
cp .env.example .env   # isi DSN PostgreSQL, URI MongoDB, Redis addr, JWT secret

# 2. Jalankan database
docker compose up -d postgres mongo redis   # setelah FT-14 ter-merge

# 3. Jalankan migration
go run ./migrations   # atau psql -f migrations/001_init.sql

# 4. Jalankan API
go run main.go

# Server di http://localhost:8080/api/v1
```

## Teknologi

Go 1.22+ · Fiber · PostgreSQL (pgx) · MongoDB · Redis + Asynq · JWT · Bcrypt · Docker Compose · Swagger (swaggo)

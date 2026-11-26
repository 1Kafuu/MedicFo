CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(80) PRIMARY KEY,
    description VARCHAR(200) NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(80) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO permissions (name, description) VALUES
    ('doctor:list', 'melihat daftar dokter'),
    ('doctor:create', 'membuat dokter'),
    ('doctor:update', 'mengubah dokter'),
    ('schedule:list', 'melihat jadwal'),
    ('schedule:create', 'membuat jadwal'),
    ('schedule:update', 'mengubah jadwal'),
    ('appointment:create', 'membuat booking'),
    ('appointment:read', 'melihat booking'),
    ('appointment:cancel', 'membatalkan booking'),
    ('appointment:status', 'mengubah status booking'),
    ('medical_record:write', 'menulis rekam medis'),
    ('medical_record:read:any', 'membaca rekam medis siapa pun'),
    ('review:create', 'membuat ulasan'),
    ('review:reply', 'membalas ulasan'),
    ('review:read', 'membaca ulasan'),
    ('report:read', 'membaca laporan'),
    ('admin:queue', 'melihat monitoring antrean'),
    ('notification:read', 'membaca notifikasi sendiri')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', name FROM permissions
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('doctor', 'doctor:list'), ('doctor', 'schedule:list'),
    ('doctor', 'appointment:read'), ('doctor', 'appointment:status'),
    ('doctor', 'medical_record:write'), ('doctor', 'medical_record:read:any'),
    ('doctor', 'review:read'), ('doctor', 'review:reply'),
    ('receptionist', 'doctor:list'), ('receptionist', 'schedule:list'),
    ('receptionist', 'appointment:read'), ('receptionist', 'appointment:status'),
    ('receptionist', 'report:read'),
    ('patient', 'doctor:list'), ('patient', 'schedule:list'),
    ('patient', 'appointment:create'), ('patient', 'appointment:read'),
    ('patient', 'appointment:cancel'), ('patient', 'review:create'),
    ('patient', 'review:read'), ('patient', 'notification:read')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dns_records (
    id SERIAL PRIMARY KEY,
    fqdn TEXT NOT NULL,
    ip TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (fqdn, ip)
);

CREATE INDEX IF NOT EXISTS idx_dns_records_fqdn ON dns_records(fqdn);
CREATE INDEX IF NOT EXISTS idx_dns_records_ip ON dns_records(ip);


CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR,
    role SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CHECK (role >= 0)
);

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE violation_types (
    id UUID PRIMARY KEY,
    code VARCHAR UNIQUE NOT NULL,
    title VARCHAR NOT NULL,
    description TEXT,
    base_fine_amount NUMERIC,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CHECK (base_fine_amount IS NULL OR base_fine_amount >= 0)
);

CREATE TABLE violation_reports (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    violation_type_id UUID NOT NULL REFERENCES violation_types(id),
    title VARCHAR NOT NULL,
    description TEXT NOT NULL,
    location TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    status SMALLINT NOT NULL,
    video_source SMALLINT NOT NULL,
    video_url TEXT,
    video_object_key TEXT,
    video_content_type TEXT,
    video_size BIGINT,
    moderator_id UUID REFERENCES users(id),
    moderation_comment TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CHECK (status >= 0),
    CHECK (video_source >= 0),
    CHECK (video_size IS NULL OR video_size >= 0)
);

CREATE TABLE payout_rules (
    id UUID PRIMARY KEY,
    violation_type_id UUID REFERENCES violation_types(id),
    percent NUMERIC NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CHECK (percent > 0 AND percent <= 100)
);

CREATE TABLE payouts (
    id UUID PRIMARY KEY,
    report_id UUID NOT NULL REFERENCES violation_reports(id),
    user_id UUID NOT NULL REFERENCES users(id),
    amount NUMERIC NOT NULL,
    status SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CHECK (amount >= 0),
    CHECK (status >= 0)
);

CREATE INDEX users_email_idx ON users(email);
CREATE INDEX user_sessions_refresh_token_hash_idx ON user_sessions(refresh_token_hash);
CREATE INDEX violation_reports_user_id_idx ON violation_reports(user_id);
CREATE INDEX violation_reports_status_idx ON violation_reports(status);
CREATE INDEX violation_reports_created_at_idx ON violation_reports(created_at);

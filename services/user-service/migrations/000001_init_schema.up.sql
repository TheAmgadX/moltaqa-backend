CREATE TABLE users (
    id UUID PRIMARY KEY,

    username TEXT NOT NULL UNIQUE,
    phone_number TEXT,
    email TEXT UNIQUE,

    profile_image_url TEXT,
    bio TEXT,
    display_name TEXT,

    email_verified TIMESTAMP NULL,
    phone_verified TIMESTAMP NULL,

    birth_date DATE,

    bio_status TEXT,

    account_badge TEXT NOT NULL DEFAULT 'UNVERIFIED',

    friends_count INTEGER NOT NULL DEFAULT 0,
    followers_count INTEGER NOT NULL DEFAULT 0,
    following_count INTEGER NOT NULL DEFAULT 0,
    posts_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE privacy_settings (
    user_id UUID PRIMARY KEY,

    avatar_visibility TEXT NOT NULL DEFAULT 'EVERYONE',
    phone_visibility TEXT NOT NULL DEFAULT 'EVERYONE',
    email_visibility TEXT NOT NULL DEFAULT 'EVERYONE',
    last_seen_visibility TEXT NOT NULL DEFAULT 'EVERYONE',

    read_receipts_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    find_by_username BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_privacy_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
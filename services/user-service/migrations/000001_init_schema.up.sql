CREATE TABLE users (
    id UUID PRIMARY KEY,

    username VARCHAR(30) NOT NULL UNIQUE,
    phone TEXT UNIQUE NULL,
    email TEXT UNIQUE NULL,

    profile_image_url TEXT NULL,
    bio VARCHAR(250) DEFAULT '',
    display_name VARCHAR(50),

    email_verified TIMESTAMP NULL,
    phone_verified TIMESTAMP NULL,

    birth_date DATE NULL,

    bio_status VARCHAR(50) DEFAULT '',

    account_badge TEXT NOT NULL DEFAULT 'unverified'
        CHECK (
            account_badge IN (
                'unverified',
                'blue_badge',
                'golden_badge',
                'silver_badge'
            )
        ),

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

    avatar_visibility TEXT NOT NULL DEFAULT 'everyone'
        CHECK (
            avatar_visibility IN (
                'everyone',
                'friends',
                'contacts',
                'nobody'
            )
        ),

    phone_visibility TEXT NOT NULL DEFAULT 'everyone'
        CHECK (
            phone_visibility IN (
                'everyone',
                'friends',
                'contacts',
                'nobody'
            )
        ),

    email_visibility TEXT NOT NULL DEFAULT 'everyone'
        CHECK (
            email_visibility IN (
                'everyone',
                'friends',
                'contacts',
                'nobody'
            )
        ),

    last_seen_visibility TEXT NOT NULL DEFAULT 'everyone'
        CHECK (
            last_seen_visibility IN (
                'everyone',
                'friends',
                'contacts',
                'nobody'
            )
        ),

    read_receipts_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    find_by_username BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_privacy_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
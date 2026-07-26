CREATE TABLE users (
    id UUID PRIMARY KEY,

    username VARCHAR(255) NOT NULL UNIQUE,
    phone_number VARCHAR(50),
    email VARCHAR(255) UNIQUE,

    profile_image_url TEXT,
    bio TEXT,
    display_name VARCHAR(255),

    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,

    birth_date DATE,

    bio_status VARCHAR(255),

    account_badge VARCHAR(50),

    friends_count INTEGER NOT NULL DEFAULT 0,
    followers_count INTEGER NOT NULL DEFAULT 0,
    following_count INTEGER NOT NULL DEFAULT 0,
    posts_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);
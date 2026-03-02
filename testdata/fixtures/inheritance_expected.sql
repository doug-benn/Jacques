CREATE TABLE users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE administrators (
    role text NOT NULL DEFAULT 'admin',
    permissions text[] NOT NULL
) INHERITS (users);

CREATE TABLE moderators (
    department text NOT NULL,
    badge_number text
) INHERITS (users);

CREATE TABLE registered_users (
    email_verified boolean NOT NULL DEFAULT false,
    last_login timestamp without time zone
) INHERITS (users);

CREATE TABLE user_sessions (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    token text NOT NULL,
    expires_at timestamp without time zone NOT NULL
);

CREATE TABLE accounts (
    id bigint PRIMARY KEY,
    username text NOT NULL
);

CREATE TABLE profiles (
    id bigint PRIMARY KEY,
    bio text
);

CREATE TABLE user_profiles (
    avatar_url text
);

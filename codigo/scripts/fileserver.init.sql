CREATE DATABASE fileserver ENCODING 'utf-8';

CREATE TABLE IF NOT EXISTS clients (
    id      VARCHAR NOT NULL PRIMARY KEY,
    secret  VARCHAR NOT NULL,
    domain  VARCHAR NOT NULL,
    user_id VARCHAR
);

CREATE TABLE IF NOT EXISTS tokens (
    client_id                       VARCHAR NOT NULL,
    user_id                         VARCHAR NOT NULL,
    redirect_uri                    VARCHAR NOT NULL,
    scope                           VARCHAR NOT NULL,
    code                            VARCHAR NOT NULL,
    code_created_at                 TIMESTAMPTZ,
    code_expires_in_seconds         INTEGER,
    code_challenge                  VARCHAR NOT NULL,
    code_challenge_method           VARCHAR NOT NULL,
    access                          VARCHAR NOT NULL,
    access_created_at               TIMESTAMPTZ,
    access_expires_in_seconds       INTEGER,
    refresh                         VARCHAR NOT NULL,
    refresh_created_at              TIMESTAMPTZ,
    refresh_expires_in_seconds      INTEGER
);

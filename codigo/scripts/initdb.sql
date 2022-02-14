/* Enable l-tree extension for paths */
CREATE EXTENSION ltree;

CREATE DATABASE indexsrv ENCODING 'utf-8';

CREATE TABLE IF NOT EXISTS organizations (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR NOT NULL PRIMARY KEY,
    name VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    access_token VARCHAR NOT NULL,
    refresh_token VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS file_servers (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    org_id INT REFERENCES organizations(id),
    name VARCHAR NOT NULL,
    auth_url VARCHAR NOT NULL,
    fetch_url VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS user_accounts (
    user_id VARCHAR NOT NULL REFERENCES users(id),
    server_id INT NOT NULL REFERENCES file_servers(id),
    token VARCHAR NOT NULL,
    refresh_token VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS mappings (
    user_id VARCHAR NOT NULL REFERENCES users(id),
    server_id INT NOT NULL REFERENCES file_servers(id),
    path ltree NOT NULL,
    ref VARCHAR NOT NULL,
    updated TIMESTAMP NOT NULL
);

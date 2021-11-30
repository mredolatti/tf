/* Enable l-tree extension for paths */
CREATE EXTENSION ltree;

CREATE TABLE IF NOT EXISTS organizations (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR NOT NULL PRIMARY KEY,
    name VARCHAR NOT NULL,
    token VARCHAR NOT NULL,
    refresh_token VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS file_servers (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    org_id INT REFERENCES organizations(id),
    name VARCHAR NOT NULL,
    auth_url VARCHAR NOT NULL,
    fetch_url VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS mappings (
    user_id VARCHAR NOT NULL REFERENCES users(id),
    server_id INT NOT NULL REFERENCES file_servers(id),
    ref VARCHAR NOT NULL,
    path ltree NOT NULL
);

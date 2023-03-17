INSERT INTO organizations(name)
VALUES
    ('organization1'),
    ('organization2'),
    ('organization3');

INSERT INTO users(name, email, password_hash)
VALUES
    ('Martín Redolatti', 'martinredolatti@gmail.com', crypt('123456', gen_salt('bf'))),
    ('user2', 'a@x.com', crypt('qwerty', gen_salt('bf'))),
    ('user3', 'b@x.com', crypt('asdfgh', gen_salt('bf')));

INSERT INTO file_servers(org_id, name, auth_url, token_url, fetch_url, control_endpoint)
VALUES
    (1, 'o1fs1', 'https://o1fs1/auth', 'https://o1fs1/token', 'https://o1fs1', 'o1fs1:1234'),
    (1, 'o1fs2', 'https://o2fs2/auth', 'https://o2fs2/token', 'https://o2fs2', 'o1fs2:1234'),
    (1, 'o1fs3', 'https://o3fs3/auth', 'https://o3fs3/token', 'https://o3fs3', 'o1fs3:1234'),
    (1, 'servercito', 'https://file-server:9877/authorize', 'https://file-server:9877/token', 'https://file-server:9877/file', 'file-server:9000');

INSERT INTO mappings(user_id, server_id, size_bytes, ref, path, updated)
VALUES
    (1, 1, 0, 'file1.jpg', 'path.to.file1__DOT__jpg', 1646394925714181390),
    (1, 1, 0, 'file2.jpg', 'path.to.file2__DOT__jpg', 1646394925714181390),
    (1, 1, 0, 'file3.jpg', 'path.another.file3__DOT__jpg', 1646394925714181390),
    (1, 1, 0, 'file4.jpg', 'path.another.file4__DOT__jpg', 1646394925714181390),
    (1, 1, 0, 'file5.jpg', 'path.yet.another.file5__DOT__jpg', 1646394925714181390),
    (2, 1, 0, 'file1.jpg', 'my.path.file1__DOT__jpg', 1646394925714181390),
    (2, 1, 0, 'file2.jpg', 'my.path.file2__DOT__jpg', 1646394925714181390),
    (3, 1, 0, 'file1.jpg', 'somewhere.file1__DOT__jpg', 1646394925714181390),
    (3, 1, 0, 'file2.jpg', 'somewhere.file2__DOT__jpg', 1646394925714181390),
    (3, 1, 0, 'file3.jpg', 'somewhere.file3__DOT__jpg', 1646394925714181390);

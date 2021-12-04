INSERT INTO organizations(name)
VALUES
    ('organization1'),
    ('organization2'),
    ('organization3');

INSERT INTO users(id, name, email, access_token, refresh_token)
VALUES
    ('107156877088323945674', 'Martín Redolatti', 'martinredolatti@gmail.com', 'qwertyuiop', 'poiuytrewq'),
    ('id2', 'user2', 'a@b.com', 'asdfghjkl', 'lkjhgfdsa'),
    ('id3', 'user3', 'a@b.com', 'zxcvbnm', 'mnbvcxz');

INSERT INTO file_servers(org_id, name, auth_url, fetch_url)
VALUES
    (1, 'o1fs1', 'https://o1fs1/auth', 'sftp://o1fs1'),
    (1, 'o1fs2', 'https://o2fs2/auth', 'sftp://o2fs2'),
    (1, 'o1fs3', 'https://o3fs3/auth', 'sftp://o3fs3');

INSERT INTO mappings(user_id, server_id, ref, path, updated)
VALUES
    ('107156877088323945674', 1, 'file1.jpg', 'path.to.file1__DOT__jpg', current_timestamp),
    ('107156877088323945674', 1, 'file2.jpg', 'path.to.file2__DOT__jpg', current_timestamp),
    ('107156877088323945674', 1, 'file3.jpg', 'path.another.file3__DOT__jpg', current_timestamp),
    ('107156877088323945674', 1, 'file4.jpg', 'path.another.file4__DOT__jpg', current_timestamp),
    ('107156877088323945674', 1, 'file5.jpg', 'path.yet.another.file5__DOT__jpg', current_timestamp),
    ('id2', 1, 'file1.jpg', 'my.path.file1__DOT__jpg', current_timestamp),
    ('id2', 1, 'file2.jpg', 'my.path.file2__DOT__jpg', current_timestamp),
    ('id3', 1, 'file1.jpg', 'somewhere.file1__DOT__jpg', current_timestamp),
    ('id3', 1, 'file2.jpg', 'somewhere.file2__DOT__jpg', current_timestamp),
    ('id3', 1, 'file3.jpg', 'somewhere.file3__DOT__jpg', current_timestamp);

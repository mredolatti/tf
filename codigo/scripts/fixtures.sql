INSERT INTO organizations(name)
VALUES
    ('organization1'),
    ('organization2'),
    ('organization3');

INSERT INTO users(id, name, email, access_token, refresh_token)
VALUES
    ('107156877088323945674', 'Martín Redolatti', 'martinredolatti@gmail.com', 'qwertyuiop', 'poiuytrewq'),
    ('id2', 'user2', 'a@x.com', 'asdfghjkl', 'lkjhgfdsa'),
    ('id3', 'user3', 'b@x.com', 'zxcvbnm', 'mnbvcxz');

INSERT INTO file_servers(id, org_id, name, auth_url, fetch_url, control_endpoint)
VALUES
    ('s1', 1, 'o1fs1', 'https://o1fs1/auth', 'sftp://o1fs1', 'o1fs1:1234'),
    ('s2', 1, 'o1fs2', 'https://o2fs2/auth', 'sftp://o2fs2', 'o1fs2:1234'),
    ('s3', 1, 'o1fs3', 'https://o3fs3/auth', 'sftp://o3fs3', 'o1fs3:1234'),
    ('fs1', 1, 'servercito', 'https://file-server/auth', 'https://file-server/file', 'file-server:9000');

INSERT INTO user_accounts(user_id, server_id, token, refresh_token, checkpoint)
VALUES
    ('107156877088323945674', 'fs1', 'none', 'none', 0);

INSERT INTO mappings(user_id, server_id, ref, path, updated)
VALUES
    ('107156877088323945674', 's1', 'file1.jpg', 'path.to.file1__DOT__jpg', 1646394925714181390),
    ('107156877088323945674', 's1', 'file2.jpg', 'path.to.file2__DOT__jpg', 1646394925714181390),
    ('107156877088323945674', 's1', 'file3.jpg', 'path.another.file3__DOT__jpg', 1646394925714181390),
    ('107156877088323945674', 's1', 'file4.jpg', 'path.another.file4__DOT__jpg', 1646394925714181390),
    ('107156877088323945674', 's1', 'file5.jpg', 'path.yet.another.file5__DOT__jpg', 1646394925714181390),
    ('id2', 's1', 'file1.jpg', 'my.path.file1__DOT__jpg', 1646394925714181390),
    ('id2', 's1', 'file2.jpg', 'my.path.file2__DOT__jpg', 1646394925714181390),
    ('id3', 's1', 'file1.jpg', 'somewhere.file1__DOT__jpg', 1646394925714181390),
    ('id3', 's1', 'file2.jpg', 'somewhere.file2__DOT__jpg', 1646394925714181390),
    ('id3', 's1', 'file3.jpg', 'somewhere.file3__DOT__jpg', 1646394925714181390);

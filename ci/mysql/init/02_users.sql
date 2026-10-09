-- Native password user for auth plugin integration tests.
CREATE USER IF NOT EXISTS 'wisp_native'@'%' IDENTIFIED WITH mysql_native_password BY 'wisp';
GRANT ALL PRIVILEGES ON wisp_bench.* TO 'wisp_native'@'%';
FLUSH PRIVILEGES;

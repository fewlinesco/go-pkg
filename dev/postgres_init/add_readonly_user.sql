CREATE USER reader_user WITH PASSWORD 'reader_user';

GRANT CONNECT ON DATABASE postgres TO reader_user;
-- This assumes you're actually connected to mydb..
ALTER DEFAULT PRIVILEGES IN SCHEMA public
   GRANT SELECT ON TABLES TO reader_user;

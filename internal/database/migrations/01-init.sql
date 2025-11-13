SELECT 'CREATE DATABASE avito'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'avito')\gexec

\c avito;

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'user') THEN
    CREATE USER "user" WITH PASSWORD 'password';
END IF;
END
$$;

GRANT ALL PRIVILEGES ON DATABASE avito TO "user";
ALTER DATABASE avito OWNER TO "user";
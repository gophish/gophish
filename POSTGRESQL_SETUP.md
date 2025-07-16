# PostgreSQL Support for Gophish

This document describes how to configure Gophish to use PostgreSQL as the database backend.

## Configuration

### 1. Update config.json

Change your `config.json` file to use PostgreSQL:

```json
{
    "admin_server": {
        "listen_url": "127.0.0.1:3333",
        "use_tls": true,
        "cert_path": "gophish_admin.crt",
        "key_path": "gophish_admin.key",
        "trusted_origins": []
    },
    "phish_server": {
        "listen_url": "0.0.0.0:80",
        "use_tls": false,
        "cert_path": "example.crt",
        "key_path": "example.key"
    },
    "db_name": "postgres",
    "db_path": "host=localhost port=5432 user=gophish password=yourpassword dbname=gophish sslmode=disable",
    "db_sslca_path": "",
    "migrations_prefix": "db/db_",
    "contact_address": "",
    "logging": {
        "filename": "",
        "level": ""
    }
}
```

### 2. PostgreSQL Connection String

The `db_path` field should contain a PostgreSQL connection string with the following format:

```
host=localhost port=5432 user=gophish password=yourpassword dbname=gophish sslmode=disable
```

Connection string parameters:
- `host`: PostgreSQL server hostname (default: localhost)
- `port`: PostgreSQL server port (default: 5432)
- `user`: Database username
- `password`: Database password
- `dbname`: Database name
- `sslmode`: SSL mode (disable, require, verify-ca, verify-full)

### 3. SSL/TLS Configuration

For SSL connections to PostgreSQL:

1. Set `sslmode=require` in the connection string
2. If using custom CA certificates, you can either:
   - Include `sslrootcert=/path/to/ca.pem` in the connection string
   - Or set `db_sslca_path` in config.json (though PostgreSQL handles SSL through connection string parameters)

Example with SSL:
```json
"db_path": "host=localhost port=5432 user=gophish password=yourpassword dbname=gophish sslmode=require sslrootcert=/path/to/ca.pem"
```

### 4. Database Setup

Before running Gophish, create the PostgreSQL database:

```sql
CREATE DATABASE gophish;
CREATE USER gophish WITH PASSWORD 'yourpassword';
GRANT ALL PRIVILEGES ON DATABASE gophish TO gophish;
```

### 5. Running Gophish

Once configured, run Gophish normally:

```bash
./gophish
```

The application will automatically detect PostgreSQL from the `db_name` setting and:
- Use the appropriate PostgreSQL driver
- Apply PostgreSQL-specific migrations from `db/db_postgres/migrations/`
- Handle the connection using the PostgreSQL connection string

## Differences from SQLite/MySQL

1. **Primary Keys**: PostgreSQL uses `SERIAL` type for auto-incrementing primary keys
2. **Data Types**: 
   - `datetime` → `timestamp`
   - `real` → `double precision`
   - Large text fields use `text` type
3. **Identifiers**: PostgreSQL doesn't require quotes around table/column names (unless they contain special characters)

## Troubleshooting

1. **Connection Refused**: Ensure PostgreSQL is running and accepting connections
2. **Authentication Failed**: Verify username/password and that the user has proper permissions
3. **SSL Errors**: Check SSL configuration and certificate paths
4. **Migration Errors**: Ensure the database user has CREATE TABLE permissions

## Existing Database Support

Gophish now supports three database backends:
- **SQLite3** (default): File-based, good for small deployments
- **MySQL**: Traditional RDBMS, good for larger deployments
- **PostgreSQL**: Advanced RDBMS with better concurrency and features
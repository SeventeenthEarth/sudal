# Database Migrations

This document describes the database migration system for the Social Quiz Platform backend, which uses `golang-migrate/migrate` for managing database schema changes.

## Overview

Database migrations provide a systematic way to manage database schema changes over time. They allow you to:

- Version control your database schema
- Apply changes consistently across environments
- Rollback changes when needed
- Collaborate safely on database changes

## Installation and Setup

The migration system is integrated into the project's Makefile and will automatically install the required tools when needed.

### Prerequisites

- PostgreSQL database (configured via environment variables)
- Go 1.24.2 or later
- Access to the database with CREATE/DROP privileges

### Environment Configuration

The migration system uses the same database configuration as the main application:

**Option 1: Using POSTGRES_DSN**

```bash
export POSTGRES_DSN="postgres://user:password@localhost:5432/quizapp_db?sslmode=disable"
```

**Option 2: Using individual components**

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=user
export DB_PASSWORD=password
export DB_NAME=quizapp_db
export DB_SSLMODE=disable
```

### Docker Development

When using Docker Compose for development, the migration commands will automatically use the containerized database:

```bash
# For Docker environment
export DB_HOST=localhost  # Docker forwards to localhost
export DB_PORT=5432
export DB_USER=user
export DB_PASSWORD=password
export DB_NAME=quizapp_db
export DB_SSLMODE=disable
```

## Migration File Structure

Migration files are stored in `db/migrations/` and follow this naming convention:

```
{version}_{description}.{direction}.sql
```

**Examples:**

- `000001_create_users_table.up.sql`
- `000001_create_users_table.down.sql`
- `000002_add_user_profiles.up.sql`
- `000002_add_user_profiles.down.sql`

### Versioning Strategy

- **Sequential numbering**: 000001, 000002, 000003, etc.
- **6-digit padding**: Ensures proper ordering
- **Descriptive names**: Use snake_case for descriptions
- **Paired files**: Each migration has both `.up.sql` and `.down.sql`

## Creating Migrations

### Using the Makefile

Create new migration files using the `migrate-create` target:

```bash
make migrate-create DESC=create_users_table
```

This will generate:

- `db/migrations/000001_create_users_table.up.sql`
- `db/migrations/000001_create_users_table.down.sql`

### Migration File Templates

**Up Migration Example (`000001_create_users_table.up.sql`):**

```sql
-- Migration: Create users table
-- Description: Initial users table for the social quiz platform
-- Version: 000001
-- Direction: UP

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Add comments for documentation
COMMENT ON TABLE users IS 'User accounts for the social quiz platform';
```

**Down Migration Example (`000001_create_users_table.down.sql`):**

```sql
-- Migration: Create users table
-- Description: Rollback the initial users table creation
-- Version: 000001
-- Direction: DOWN

-- Drop indexes first
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_email;

-- Drop the users table
DROP TABLE IF EXISTS users;
```

## Applying Migrations

### Apply All Pending Migrations

```bash
make migrate-up
```

This command:

- Installs the migrate CLI if not present
- Applies all pending migrations in order
- Shows progress and results

### Check Migration Status

```bash
make migrate-status
```

Shows:

- Current migration version
- Available migration files
- Database connection status

### Show Current Version

```bash
make migrate-version
```

Displays the current migration version number.

## Database Reset and Fresh Setup

### Reset Database (Keep Migration Files)

```bash
make migrate-reset
```

This command:

- Drops ALL database objects (tables, views, functions, triggers, sequences, types, etc.)
- Uses `DROP SCHEMA CASCADE` for complete cleanup
- Reapplies all migrations from scratch
- Keeps existing migration files unchanged
- Useful for testing migration sequences

### Fresh Migration Setup

```bash
make migrate-fresh
```

**⚠️ WARNING: This removes all migration files!**

This command:

- Creates backup of existing migration files
- Drops ALL database objects using `DROP SCHEMA CASCADE`
- Clears migration files directory
- Allows starting with completely fresh migrations

### Drop All Tables

```bash
make migrate-drop
```

**⚠️ DANGER: This deletes all data!**

This command:

- Drops ALL database objects using `DROP SCHEMA CASCADE`
- Includes tables, views, functions, triggers, sequences, types, etc.
- Resets migration state completely
- Requires typing "DROP ALL DATA" to confirm

### Use Cases for Reset Commands

**`migrate-reset`** - When you want to:

- Test the complete migration sequence
- Fix migration order issues
- Verify all migrations work from scratch
- Keep existing migration files

**`migrate-fresh`** - When you want to:

- Start completely over with new migrations
- Clean up messy development history
- Create a single initial migration
- Remove experimental migrations

**`migrate-drop`** - When you want to:

- Manually clean the database
- Prepare for manual schema setup
- Emergency database cleanup

### Why DROP SCHEMA CASCADE?

The migration system uses `DROP SCHEMA CASCADE` instead of just dropping tables because:

**Complete Cleanup**: Removes ALL database objects in one operation:

- Tables and their data
- Views (materialized and regular)
- Functions and stored procedures
- Triggers and trigger functions
- Sequences and serial columns
- Custom types (ENUMs, composite types)
- Indexes (including partial and functional indexes)
- Constraints (foreign keys, check constraints, etc.)
- Comments and metadata

**Dependency Resolution**: CASCADE automatically handles object dependencies, preventing errors like:

- "cannot drop table because view depends on it"
- "cannot drop function because trigger depends on it"
- "cannot drop type because table column depends on it"

**Guaranteed Clean State**: Ensures no orphaned objects remain that could interfere with fresh migrations.

## Rolling Back Migrations

### Rollback Last Migration

```bash
make migrate-down
```

**⚠️ Warning**: This will prompt for confirmation before rolling back.

### Force Set Migration Version

For recovery scenarios only:

```bash
make migrate-force
```

**⚠️ Danger**: This is a dangerous operation that should only be used when the migration state is corrupted.

## Best Practices

### Writing Migrations

1. **Always write both up and down migrations**
2. **Use transactions when possible** (PostgreSQL supports DDL transactions)
3. **Include comments** explaining the purpose
4. **Test migrations** on a copy of production data
5. **Keep migrations small** and focused on one change
6. **Use IF EXISTS/IF NOT EXISTS** for safety

### Migration Safety

1. **Backup before major changes**
2. **Test rollbacks** before applying to production
3. **Avoid destructive changes** in up migrations when possible
4. **Use feature flags** for application changes that depend on schema changes

### Naming Conventions

- Use descriptive names: `add_user_profiles`, `create_quiz_tables`
- Use snake_case for consistency
- Include the action: `create_`, `add_`, `remove_`, `modify_`
- Be specific: `add_email_index` not `add_index`

## Troubleshooting

### Common Issues

**Migration fails with "dirty database":**

```bash
# Check current version and fix manually
make migrate-version
make migrate-force  # Use with caution
```

**Database connection issues:**

```bash
# Verify environment variables
echo $DATABASE_URL
# Or check individual components
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME
```

**Migration files not found:**

```bash
# Ensure migrations directory exists
ls -la db/migrations/
# Create if missing
make migrate-create DESC=initial_setup
```

### Recovery Procedures

1. **Backup your database** before attempting recovery
2. **Identify the issue** using `make migrate-status`
3. **Check migration logs** for specific errors
4. **Use migrate-force** only as a last resort
5. **Verify data integrity** after recovery

## Integration with Application

The migration system is designed to work alongside the existing database connection pooling system. Migrations should be applied before starting the application in production environments.

### Recommended Deployment Flow

1. **Backup database**
2. **Apply migrations**: `make migrate-up`
3. **Verify migration status**: `make migrate-status`
4. **Start application**
5. **Verify application functionality**

## Migration Squashing

**⚠️ WARNING: Only use in development environments!**

Migration squashing combines multiple migration files into a single migration. This is useful for:

- Cleaning up development history
- Reducing the number of migration files
- Creating a clean starting point

### When to Squash

- ✅ **Development environment only**
- ✅ **Before first production deployment**
- ✅ **When you have many small migrations**
- ❌ **Never in production**
- ❌ **Never when other developers have applied migrations**

### How to Squash

```bash
# 1. Run the squashing tool
make migrate-squash

# 2. Create your squashed migration manually
# Use the schema dump at /tmp/schema_dump.sql as reference

# 3. Test the squashed migration
DB_NAME=test_squashed_migration make migrate-up

# 4. Compare schemas to ensure they match
# 5. Replace old migrations with squashed ones
# 6. Reset migration version
```

### Manual Squashing Process

1. **Backup existing migrations**
2. **Create schema dump** of current database
3. **Create new migration file** with complete schema
4. **Test on fresh database**
5. **Replace old migration files**
6. **Reset migration version**

## Additional Resources

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Migration Best Practices](https://www.postgresql.org/docs/current/ddl-alter.html)
- [Database Migration Strategies](https://martinfowler.com/articles/evodb.html)

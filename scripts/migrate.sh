#!/bin/bash

# Database Migration Management Script
# This script manages all database migration operations for the Sudal project
# Extracted from Makefile to improve maintainability and reduce complexity

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration - can be overridden by environment variables
MIGRATIONS_DIR="${MIGRATIONS_DIR:-./db/migrations}"
DEFAULT_DB_HOST="${DEFAULT_DB_HOST:-localhost}"
DEFAULT_DB_PORT="${DEFAULT_DB_PORT:-5432}"
DEFAULT_DB_USER="${DEFAULT_DB_USER:-user}"
DEFAULT_DB_PASSWORD="${DEFAULT_DB_PASSWORD:-password}"
DEFAULT_DB_NAME="${DEFAULT_DB_NAME:-quizapp_db}"
DEFAULT_DB_SSLMODE="${DEFAULT_DB_SSLMODE:-disable}"
DEFAULT_DB_SCHEMA="${DEFAULT_DB_SCHEMA:-sudal}"

# Project schemas - add new schemas here as they are introduced
PROJECT_SCHEMAS="sudal"  # Space-separated list of schemas to manage

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to handle errors
handle_error() {
    print_error "$1"
    exit 1
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to clean environment variable values (remove comments)
clean_env_var() {
    echo "$1" | cut -d' ' -f1
}

# Function to get database schema
get_database_schema() {
    local db_schema
    db_schema=$(clean_env_var "${DB_SCHEMA:-}")
    echo "${db_schema:-$DEFAULT_DB_SCHEMA}"
}

# Function to get project schemas (can be overridden by environment variable)
get_project_schemas() {
    local schemas
    schemas="${PROJECT_SCHEMAS_OVERRIDE:-$PROJECT_SCHEMAS}"
    echo "$schemas"
}

# Function to drop all project schemas
drop_all_project_schemas() {
    local database_url schemas
    database_url="$1"
    schemas=$(get_project_schemas)

    print_info "Dropping all project schemas: $schemas"

    for schema in $schemas; do
        print_info "Dropping schema '$schema' and all its objects..."
        psql "$database_url" -c "DROP SCHEMA IF EXISTS $schema CASCADE;" || handle_error "Failed to drop schema '$schema'"
        print_success "Schema '$schema' dropped successfully"
    done
}

# Function to create all project schemas
create_all_project_schemas() {
    local database_url schemas
    database_url="$1"
    schemas=$(get_project_schemas)

    print_info "Creating all project schemas: $schemas"

    for schema in $schemas; do
        print_info "Creating schema '$schema'..."
        psql "$database_url" -c "CREATE SCHEMA IF NOT EXISTS $schema;" || handle_error "Failed to create schema '$schema'"
        print_success "Schema '$schema' created successfully"
    done
}

# Function to construct DATABASE_URL
construct_database_url() {
    local db_host db_port db_user db_password db_name db_sslmode

    # Use POSTGRES_DSN if available
    if [ -n "$POSTGRES_DSN" ]; then
        echo "$POSTGRES_DSN"
        return 0
    fi

    # Clean environment variables (remove comments)
    db_host=$(clean_env_var "${DB_HOST:-}")
    db_port=$(clean_env_var "${DB_PORT:-}")
    db_user=$(clean_env_var "${DB_USER:-}")
    db_password=$(clean_env_var "${DB_PASSWORD:-}")
    db_name=$(clean_env_var "${DB_NAME:-}")
    db_sslmode=$(clean_env_var "${DB_SSLMODE:-}")

    # Apply defaults and handle special cases
    # Note: Use localhost for local development, even when Docker Compose uses 'db' internally
    if [ "$db_host" = "db" ]; then
        db_host="localhost"
    fi

    db_host="${db_host:-$DEFAULT_DB_HOST}"
    db_port="${db_port:-$DEFAULT_DB_PORT}"
    db_user="${db_user:-$DEFAULT_DB_USER}"
    db_password="${db_password:-$DEFAULT_DB_PASSWORD}"
    db_name="${db_name:-$DEFAULT_DB_NAME}"
    db_sslmode="${db_sslmode:-$DEFAULT_DB_SSLMODE}"

    # Construct DATABASE_URL
    echo "postgres://${db_user}:${db_password}@${db_host}:${db_port}/${db_name}?sslmode=${db_sslmode}"
}

# Function to ensure migrate tool is installed
ensure_migrate_tool() {
    if ! command_exists migrate; then
        print_warning "migrate CLI not found. Installing..."
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest || handle_error "Failed to install migrate CLI"
        print_success "migrate CLI installed successfully."
    fi
}

# Function to get migrate command (handle both system and GOPATH installations)
get_migrate_cmd() {
    if command_exists migrate; then
        echo "migrate"
    else
        echo "$(go env GOPATH)/bin/migrate"
    fi
}

# Function to ensure migrations directory exists
ensure_migrations_dir() {
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        print_error "Migrations directory $MIGRATIONS_DIR does not exist"
        print_info "Run '$0 create DESC=initial_setup' to create your first migration"
        exit 1
    fi
}

# Function to apply migrations (up)
migrate_up() {
    local database_url
    database_url=$(construct_database_url)

    print_info "=== Applying database migrations ==="
    print_info "Database URL: $database_url"
    print_info "Migrations directory: $MIGRATIONS_DIR"

    ensure_migrate_tool
    ensure_migrations_dir

    print_info "Running migrations..."
    local migrate_cmd
    migrate_cmd=$(get_migrate_cmd)

    $migrate_cmd -path "$MIGRATIONS_DIR" -database "$database_url" up || handle_error "Migration failed"

    print_success "=== Database migrations applied successfully ==="
}

# Function to rollback migrations (down)
migrate_down() {
    local database_url steps
    database_url=$(construct_database_url)
    steps="${1:-1}"

    print_info "=== Rolling back database migrations ==="
    print_info "Database URL: $database_url"
    print_info "Migrations directory: $MIGRATIONS_DIR"
    print_warning "This will rollback the last $steps migration(s)!"

    echo "Press Ctrl+C to cancel, or Enter to continue..."
    read -r dummy

    ensure_migrate_tool

    local migrate_cmd
    migrate_cmd=$(get_migrate_cmd)

    $migrate_cmd -path "$MIGRATIONS_DIR" -database "$database_url" down "$steps" || handle_error "Migration rollback failed"

    print_success "=== Migration(s) rolled back successfully ==="
}

# Function to show migration status
migrate_status() {
    local database_url
    database_url=$(construct_database_url)

    print_info "=== Database migration status ==="
    print_info "Database URL: $database_url"
    print_info "Migrations directory: $MIGRATIONS_DIR"

    ensure_migrate_tool

    if [ ! -d "$MIGRATIONS_DIR" ]; then
        print_warning "Migrations directory $MIGRATIONS_DIR does not exist"
        return 1
    fi

    print_info "Current migration version:"
    local migrate_cmd
    migrate_cmd=$(get_migrate_cmd)

    $migrate_cmd -path "$MIGRATIONS_DIR" -database "$database_url" version || print_info "No migrations applied yet"

    echo ""
    print_info "Available migration files:"
    ls -la "$MIGRATIONS_DIR"/ 2>/dev/null || print_info "No migration files found"
}

# Function to show current migration version
migrate_version() {
    local database_url
    database_url=$(construct_database_url)

    print_info "=== Current migration version ==="

    ensure_migrate_tool

    local migrate_cmd
    migrate_cmd=$(get_migrate_cmd)

    $migrate_cmd -path "$MIGRATIONS_DIR" -database "$database_url" version
}

# Function to force set migration version
migrate_force() {
    local database_url version
    database_url=$(construct_database_url)

    print_info "=== Force set migration version ==="
    print_warning "This is a dangerous operation that should only be used for recovery!"

    print_info "Current version:"
    migrate_version || print_warning "Could not determine current version"

    echo ""
    echo "Enter the version number to force set (or Ctrl+C to cancel):"
    read -r version

    if [ -z "$version" ]; then
        print_error "No version provided. Cancelling."
        exit 1
    fi

    print_info "Setting migration version to $version..."

    ensure_migrate_tool

    local migrate_cmd
    migrate_cmd=$(get_migrate_cmd)

    $migrate_cmd -path "$MIGRATIONS_DIR" -database "$database_url" force "$version" || handle_error "Failed to force migration version"

    print_success "=== Migration version forced successfully ==="
}

# Function to create new migration files
migrate_create() {
    local description="$1"

    if [ -z "$description" ]; then
        print_error "DESC parameter is required"
        print_info "Usage: $0 create DESC=create_users_table"
        exit 1
    fi

    print_info "=== Creating new migration files ==="

    ensure_migrate_tool

    mkdir -p "$MIGRATIONS_DIR"

    print_info "Creating migration files for: $description"

    local migrate_cmd
    migrate_cmd=$(get_migrate_cmd)

    $migrate_cmd create -ext sql -dir "$MIGRATIONS_DIR" -seq "$description" || handle_error "Failed to create migration files"

    print_success "=== Migration files created successfully ==="
    print_info "Edit the generated .up.sql and .down.sql files in $MIGRATIONS_DIR/"
}

# Function to reset database (drop all project schemas and reapply all migrations)
migrate_reset() {
    local database_url schemas
    database_url=$(construct_database_url)
    schemas=$(get_project_schemas)

    print_info "=== Resetting database ==="
    print_info "Database URL: $database_url"
    print_info "Project schemas: $schemas"
    print_info "Migrations directory: $MIGRATIONS_DIR"
    print_warning "This will drop ALL database objects in ALL project schemas and reapply migrations!"

    echo "Press Ctrl+C to cancel, or Enter to continue..."
    read -r dummy

    ensure_migrate_tool
    ensure_migrations_dir

    # Drop and recreate all project schemas
    drop_all_project_schemas "$database_url"
    create_all_project_schemas "$database_url"

    print_info "Reapplying all migrations..."
    local migrate_cmd
    migrate_cmd=$(get_migrate_cmd)

    $migrate_cmd -path "$MIGRATIONS_DIR" -database "$database_url" up || handle_error "Migration failed"

    print_success "=== Database reset completed successfully ==="
}

# Function to drop all database objects (dangerous operation)
migrate_drop() {
    local database_url schemas
    database_url=$(construct_database_url)
    schemas=$(get_project_schemas)

    print_info "=== Dropping all database objects ==="
    print_info "Database URL: $database_url"
    print_info "Project schemas: $schemas"
    print_warning "This will DELETE ALL DATA and ALL OBJECTS in ALL project schemas!"
    print_warning "This includes tables, views, functions, triggers, sequences, types, etc."

    echo "Type 'DROP ALL DATA' to confirm:"
    read -r confirm

    if [ "$confirm" != "DROP ALL DATA" ]; then
        print_info "Operation cancelled."
        exit 0
    fi

    ensure_migrate_tool

    # Drop all project schemas
    drop_all_project_schemas "$database_url"

    print_success "=== All database objects in ALL project schemas dropped successfully ==="
}

# Function to fresh migration setup (backup migrations and start clean)
migrate_fresh() {
    local database_url schemas backup_dir
    database_url=$(construct_database_url)
    schemas=$(get_project_schemas)
    backup_dir="./db/migrations_backup_$(date +%Y%m%d_%H%M%S)"

    print_info "=== Fresh migration setup ==="
    print_info "Database URL: $database_url"
    print_info "Project schemas: $schemas"
    print_info "Migrations directory: $MIGRATIONS_DIR"
    print_warning "This will backup and remove all current migration files!"
    print_warning "This will also drop ALL database objects in ALL project schemas!"

    echo "Press Ctrl+C to cancel, or Enter to continue..."
    read -r dummy

    ensure_migrate_tool

    # Backup existing migrations if they exist
    if [ -d "$MIGRATIONS_DIR" ] && [ "$(ls -A "$MIGRATIONS_DIR" 2>/dev/null)" ]; then
        print_info "Backing up existing migrations to $backup_dir..."
        mkdir -p "$backup_dir"
        cp -r "$MIGRATIONS_DIR"/* "$backup_dir"/ || handle_error "Failed to backup migrations"
        print_success "Migrations backed up to $backup_dir"
    fi

    # Drop and recreate all project schemas
    drop_all_project_schemas "$database_url"
    create_all_project_schemas "$database_url"

    # Clear migrations directory
    print_info "Clearing migrations directory..."
    rm -rf "$MIGRATIONS_DIR"/*

    print_success "=== Fresh migration setup completed ==="
    print_info "You can now create new migrations with: $0 create DESC=initial_setup"
    if [ -d "$backup_dir" ]; then
        print_info "Old migrations are backed up in: $backup_dir"
    fi
}

# Function to display help
show_help() {
    echo "Usage: $0 COMMAND [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  up [STEPS]        Apply pending migrations (default: all)"
    echo "  down [STEPS]      Rollback migrations (default: 1)"
    echo "  status            Show migration status"
    echo "  version           Show current migration version"
    echo "  force VERSION     Force set migration version (dangerous!)"
    echo "  create DESC       Create new migration files"
    echo "  reset             Reset database and reapply all migrations"
    echo "  drop              Drop all database objects (DANGEROUS!)"
    echo "  fresh             Fresh migration setup - backup and start clean"
    echo "  help              Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  POSTGRES_DSN              Complete database connection string"
    echo "  DB_HOST                   Database host (default: $DEFAULT_DB_HOST)"
    echo "  DB_PORT                   Database port (default: $DEFAULT_DB_PORT)"
    echo "  DB_USER                   Database user (default: $DEFAULT_DB_USER)"
    echo "  DB_PASSWORD               Database password (default: $DEFAULT_DB_PASSWORD)"
    echo "  DB_NAME                   Database name (default: $DEFAULT_DB_NAME)"
    echo "  DB_SSLMODE                SSL mode (default: $DEFAULT_DB_SSLMODE)"
    echo "  DB_SCHEMA                 Database schema for migrations (default: $DEFAULT_DB_SCHEMA)"
    echo "  PROJECT_SCHEMAS_OVERRIDE  Override project schemas (default: $PROJECT_SCHEMAS)"
    echo "  MIGRATIONS_DIR            Migrations directory (default: $MIGRATIONS_DIR)"
    echo ""
    echo "Examples:"
    echo "  $0 up                           # Apply all pending migrations"
    echo "  $0 down                         # Rollback last migration"
    echo "  $0 down 3                       # Rollback last 3 migrations"
    echo "  $0 status                       # Show migration status"
    echo "  $0 create initial_schema        # Create new migration"
    echo "  $0 force 5                      # Force set version to 5"
    echo "  $0 reset                        # Reset database and reapply migrations"
    echo "  $0 drop                         # Drop all database objects"
    echo "  $0 fresh                        # Fresh migration setup"
    echo ""
    echo "Schema Management:"
    echo "  Current project schemas: $PROJECT_SCHEMAS"
    echo "  To add more schemas, update PROJECT_SCHEMAS in this script"
    echo "  Example: PROJECT_SCHEMAS_OVERRIDE=\"sudal temp org\" $0 reset"
    echo ""
    echo "Note: If POSTGRES_DSN is set, it takes precedence over individual DB_* variables."
    echo "Note: reset/drop/fresh operations affect ALL project schemas, not just the default."
}

# Main function
main() {
    local command="$1"
    shift || true

    case "$command" in
        "up")
            migrate_up "$@"
            ;;
        "down")
            migrate_down "$@"
            ;;
        "status")
            migrate_status
            ;;
        "version")
            migrate_version
            ;;
        "force")
            migrate_force "$@"
            ;;
        "create")
            migrate_create "$@"
            ;;
        "reset")
            migrate_reset
            ;;
        "drop")
            migrate_drop
            ;;
        "fresh")
            migrate_fresh
            ;;
        "help"|"--help"|"-h"|"")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            print_info "Use '$0 help' for usage information."
            exit 1
            ;;
    esac
}

# Run main function
main "$@"

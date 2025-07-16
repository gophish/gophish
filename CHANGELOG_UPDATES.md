# Changelog - UI Fixes and PostgreSQL Support

## Date: 2025-07-16

### 1. Fixed Email Template UI Issues

#### Issue
Users experienced UI glitches when adding email templates:
- Screen would glitch out
- Couldn't add valid templates
- Scrolling issues in the modal

#### Root Causes Identified
1. CKEditor wasn't properly synchronized when toggling the "Add Tracking Image" checkbox
2. Modal scrolling fix was incomplete
3. Template validation was too strict with non-Go template syntax

#### Fixes Applied

**File: `static/js/src/app/templates.js`**
- Added real-time synchronization between tracking checkbox and CKEditor content
- Implemented proper error handling for templates without `</body>` tags
- Enhanced modal scrolling fix with proper scrollbar width calculation
- Added modal body max-height and overflow handling

**File: `models/template_context.go`**
- Fixed non-constant format string linter warning in error handling

### 2. Added PostgreSQL Database Support

#### Features Added
- Full PostgreSQL support alongside existing SQLite3 and MySQL support
- SSL/TLS connection support for PostgreSQL
- Complete migration files for PostgreSQL

#### Implementation Details

**File: `models/models.go`**
- Added PostgreSQL driver imports (`github.com/jinzhu/gorm/dialects/postgres` and `github.com/lib/pq`)
- Updated `chooseDBDriver` function to handle PostgreSQL dialect
- Added PostgreSQL SSL/TLS configuration support (via connection string)

**Migration Files: `db/db_postgres/migrations/`**
- Created complete set of PostgreSQL-compatible migrations
- Converted all 25 migration files from SQLite format to PostgreSQL format
- Key conversions:
  - `integer primary key autoincrement` → `SERIAL PRIMARY KEY`
  - `datetime` → `timestamp`
  - `real` → `double precision`
  - Text fields properly sized

**Documentation: `POSTGRESQL_SETUP.md`**
- Comprehensive setup guide for PostgreSQL
- Connection string format and parameters
- SSL/TLS configuration instructions
- Database setup commands
- Troubleshooting guide

### 3. Configuration Changes

Users can now configure PostgreSQL in `config.json`:
```json
{
    "db_name": "postgres",
    "db_path": "host=localhost port=5432 user=gophish password=yourpassword dbname=gophish sslmode=disable"
}
```

### Summary of Modified Files
1. `static/js/src/app/templates.js` - UI fixes for template editing
2. `models/template_context.go` - Fixed linter warning
3. `models/models.go` - Added PostgreSQL support
4. `db/db_postgres/` - New directory with migrations
5. `POSTGRESQL_SETUP.md` - New documentation file
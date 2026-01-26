# Database Setup

This directory contains all database-related files for the Room Draw application.

## Directory Structure

```
database/
├── scripts/        # Python scripts for database operations
├── sql/            # SQL table creation and management scripts
├── dorms/          # JSON configuration files for each dorm
├── data/           # CSV data files (numbers, preplacements, collisions)
├── notebooks/      # Jupyter notebooks for data manipulation and testing
├── .env            # Environment variables (not in version control)
├── .env.example    # Template for .env file
└── requirements.txt
```

## Prerequisites

1. PostgreSQL database server
2. Python 3.x with pip
3. Required Python packages: `pip install -r requirements.txt`

## Environment Setup

1. Copy `.env.example` to `.env`
2. Fill in your database credentials:
   - `SQL_PASS` - Database password
   - `SQL_IP` - Database server IP/hostname
   - `SQL_DB_NAME` - Database name
   - `SQL_USER` - Database username
   - `SQL_PORT` - Database port (optional)
   - `USE_SSL` - Whether to use SSL connection

## Usage

### Initial Setup

Run these commands from the `scripts/` directory:

```bash
cd scripts

# 1. Create all database tables (drops existing first!)
python createAllTables.py

# 2. Populate dorm/room data from JSON files
python createDorms.py
```

### Data Population (for testing)

```bash
cd scripts
python insertUserData.py  # Generate fake user data
```

### Production Data Import

Use the Jupyter notebooks in `notebooks/` directory. **Note:** Run notebooks from the `database/` directory (not from `notebooks/`) so that `.env` paths resolve correctly.

- `insertNumbers.ipynb` - Import draw numbers from CSV
- `insertPreplacements.ipynb` - Import preplaced students
- `insertInDorm.ipynb` - Import in-dorm status
- `insertGenderPreference.ipynb` - Import gender preferences
- `FakePopulate.ipynb` - Generate fake/test data

## File Descriptions

### SQL Files (`sql/`)
| File | Description |
|------|-------------|
| `CreateUserTable.sql` | Users table schema |
| `CreateRoomTable.sql` | Rooms table schema |
| `CreateSuitesTable.sql` | Suites table schema |
| `CreateGroupsTable.sql` | Suite groups table schema |
| `CreateRateLimitTable.sql` | Rate limiting table schema |
| `CreateTransactionLogsTable.sql` | Transaction logging table schema |
| `DropTables.sql` | Drop all tables (use with caution!) |

### Dorm JSON Files (`dorms/`)
Each JSON file contains the floor/suite/room configuration for a dorm:
- `atwood.json`, `case.json`, `drinkward.json`, `east.json`
- `linde.json`, `north.json`, `sontag.json`, `south.json`, `west.json`

### Scripts (`scripts/`)
| Script | Description |
|--------|-------------|
| `createAllTables.py` | Creates all database tables (drops existing first) |
| `createDorms.py` | Populates dorm/room data from JSON files |
| `insertUserData.py` | Generates fake user data for testing |
| `checkcollision.py` | Checks for collisions between numbers and preplacements |

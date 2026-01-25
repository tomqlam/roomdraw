# Digital Draw App for Harvey Mudd College's Room Draw

This app facilitates the Digital Draw process at Harvey Mudd, encoding all of the rules for Room Draw for pulling rooms.

## TL;DR - Fastest Setup (Using Staging Database)

If you just want to get running quickly using the shared staging database:

```bash
# 1. Setup frontend
cd frontend && npm install

# 2. Setup backend env (get credentials from team lead)
cp backend/.env.example backend/.env
# Then edit backend/.env and fill in the values

# 3. Start backend (in one terminal)
cd backend
docker build -t roomdraw-backend .
docker run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend

# 4. Start frontend (in another terminal)
cd frontend && npm start

# 5. Open http://localhost:3000
```

> **Note:** Ask the team lead for staging database and BunnyNet credentials. For local development with your own database, see the full setup below.

---

## Prerequisites

- **Docker** - For running the backend server
- **Node.js** (v18+) - For the React frontend
- **Python 3.12+** - For database setup scripts
- **PostgreSQL 14+** - For local development database

### Installation Guides

#### Docker
- [Docker Desktop Documentation](https://docs.docker.com/desktop/) - Main documentation
- [Install on Windows](https://docs.docker.com/desktop/setup/install/windows-install/)
- [Install on Mac](https://docs.docker.com/desktop/setup/install/mac-install/)
- [Install on Linux](https://docs.docker.com/desktop/setup/install/linux/)

#### Node.js
- [Official Downloads](https://nodejs.org/en/download) - Download installers
- [Installing via Package Manager](https://nodejs.org/en/download/package-manager/all) - Recommended for most users
- [npm Documentation](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm/) - Includes nvm (version manager) instructions

#### Python
- [Official Downloads](https://www.python.org/downloads/) - Download installers
- [Beginner's Guide](https://wiki.python.org/moin/BeginnersGuide/Download) - Platform-specific instructions
- [Real Python Installation Guide](https://realpython.com/installing-python/) - Detailed walkthrough

#### PostgreSQL
- [Official Downloads](https://www.postgresql.org/download/) - All platforms
- [Postgres.app](https://postgresapp.com/) - Easiest option for Mac
- [Installation Tutorial](https://www.postgresql.org/docs/current/tutorial-install.html) - Official documentation

#### macOS Users (Homebrew)
[Homebrew](https://brew.sh/) simplifies installing most dependencies on Mac:
```bash
# Install Homebrew first
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Then install dependencies
brew install node python postgresql@14
brew install --cask docker
```
See [Homebrew Installation Docs](https://docs.brew.sh/Installation) for details.

## Quick Start

1. Set up the database (see Database Setup below)
2. Configure environment files (see Environment Configuration below)
3. Start the backend: `cd backend && docker build -t roomdraw-backend . && docker run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend`
4. Start the frontend: `cd frontend && npm install && npm start`
5. Open http://localhost:3000

## Environment Configuration

### Backend Environment File

The backend looks for `backend/.env`. Copy the template and fill in your values:

```bash
cp backend/.env.example backend/.env
```

**For local development** (create `backend/.env`):

```bash
# Database connection
SQL_PASS="your_local_postgres_password"
SQL_IP="host.docker.internal"    # For Docker to connect to host machine's PostgreSQL
SQL_USER="postgres"
SQL_DB_NAME="roomdraw"
SQL_PORT="5432"
USE_SSL="disable"                # Use "disable" for local, "require" for cloud/production

# Authentication
REQUIRE_AUTH="False"             # Set to "False" to bypass Google OAuth during development

# BunnyNet CDN (required for suite design images)
BUNNYNET_WRITE_API_KEY="<get from team lead>"
BUNNYNET_READ_API_KEY="<get from team lead>"
BUNNYNET_STORAGE_ZONE="digidraw-production"
CDN_URL="https://digitaldraw.b-cdn.net"

# Email notifications (optional for local dev)
EMAIL_USERNAME=""
EMAIL_PASSWORD=""
```

**For staging/production** (see `backend/.env.production`):

```bash
# Database - connects to HMC's server
SQL_PASS="<password>"
SQL_IP="ark.cs.hmc.edu"
SQL_USER="roomdraw24"
SQL_DB_NAME="roomdraw24"
SQL_PORT="5432"
USE_SSL="require"

# Authentication enabled
REQUIRE_AUTH="True"

# BunnyNet CDN
BUNNYNET_WRITE_API_KEY="<api-key>"
BUNNYNET_READ_API_KEY="<api-key>"
BUNNYNET_STORAGE_ZONE="digidraw-production"
CDN_URL="https://digitaldraw.b-cdn.net"

# Email notifications (HMC SMTP)
EMAIL_USERNAME="<hmc-username>"
EMAIL_PASSWORD="<password>"
```

### Database Environment File

The database scripts look for `database/.env`. Copy the template and fill in your values:

```bash
cp database/.env.example database/.env
```

Variables:
| Variable | Description |
|----------|-------------|
| `SQL_PASS` | Database password |
| `SQL_IP` | Database host (`localhost` for local, IP for cloud) |
| `SQL_DB_NAME` | Database name |
| `SQL_USER` | Database username |

> **Note:** The database scripts use `SQL_*` variable names. Make sure your `.env` matches this format.

## Database Setup

### Local PostgreSQL Setup

1. Install PostgreSQL and create a database:
   ```bash
   createdb roomdraw
   ```

2. Install Python dependencies:
   ```bash
   cd database
   pip install -r requirements.txt
   ```

3. Configure `database/.env` with your local PostgreSQL credentials

4. Create all tables:
   ```bash
   python createAllTables.py
   ```

5. Populate dorm data:
   ```bash
   python createDorms.py
   ```

6. (Optional) Populate test user data using the Jupyter notebooks:
   - `FakePopulate.ipynb` - Creates fake test users
   - `insertNumbers.ipynb` - Assigns draw numbers
   - `insertPreplacements.ipynb` - Sets up preplaced users

   To run notebooks, install Jupyter: `pip install jupyter` then `jupyter notebook`
   See [Installing Jupyter](https://jupyter.org/install) for more details.

### Database Tables

The setup scripts create the following tables in order:
1. **Suites** - Dorm suites (groups of rooms)
2. **SuiteGroups** - Groups of users pulling together
3. **Users** - Student information and draw numbers
4. **Rooms** - Individual rooms within suites
5. **user_rate_limits** - Rate limiting and blacklist tracking
6. **transaction_logs** - Audit log for room changes

## External Services Setup

### BunnyNet CDN (Required)

BunnyNet is used for storing suite design images uploaded by students.

1. Create a BunnyNet account at https://bunny.net
2. Create a Storage Zone from the Storage tab in the dashboard
   - Choose a main storage region closest to your users
   - Optionally add replication regions (cannot be removed later)
3. Get your API keys: Go to Storage Zone → FTP & API Access → copy the Password
4. Create a Pull Zone connected to your Storage Zone:
   - Go to CDN → Add Pull Zone → Select "Storage Zone" as origin type
   - This gives you the CDN URL (e.g., `https://your-zone.b-cdn.net`)
5. Add all values to your backend `.env` file

**Resources:**
- [How to create a Pull Zone](https://support.bunny.net/hc/en-us/articles/207790269-How-to-create-your-first-Pull-Zone)
- [How to access files from Bunny Storage](https://support.bunny.net/hc/en-us/articles/8561433879964-How-to-access-and-deliver-files-from-Bunny-Storage)

### Google Cloud SQL (For Staging)

For shared staging environments with a cloud-hosted PostgreSQL database:

1. Enable the Cloud SQL Admin API in Google Cloud Console
2. Go to Cloud SQL → Create Instance → Choose PostgreSQL
3. Configure instance settings:
   - Choose a region close to your users (cannot be changed later)
   - Set a password for the `postgres` user
4. Configure authorized networks to allow connections from developer IPs
5. Create a database and user for the application
6. Use the instance's public IP and credentials in your `.env`

**Resources:**
- [Cloud SQL for PostgreSQL Documentation](https://docs.cloud.google.com/sql/docs/postgres)
- [Create Instances Guide](https://docs.cloud.google.com/sql/docs/postgres/create-instance)
- [Production Setup Guide](https://cloud.google.com/solutions/setting-up-cloud-sql-for-postgresql-for-production)

### HMC SMTP (For Notifications)

Email notifications use Harvey Mudd's SMTP server. Contact the HMC CS department for SMTP credentials if you need to test email functionality.

Currently, we're just using my (tlam@g.hmc.edu) email for notifications, which is kinda scuffed. You can edit the .production.env file to use your own email username and password.

## Running the Application

### Backend (Go + Docker)

```bash
cd backend

# Development with live reload:
docker build -t roomdraw-backend .
docker run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
```

The backend runs on http://localhost:8080

### Frontend (React)

```bash
cd frontend
npm install
npm start
```

The frontend runs on http://localhost:3000

### Production Deployment

For production, use Podman and the production environment file:

```bash
cd backend
podman build -t roomdraw-backend --build-arg ENV=production .
podman run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
```

## Project Structure

```
roomdraw/
├── backend/           # Go backend (Gin framework)
│   ├── cmd/server/    # Main entry point
│   ├── pkg/
│   │   ├── handlers/  # API route handlers
│   │   ├── middleware/# Auth, rate limiting, request queue
│   │   ├── models/    # Database models and types
│   │   ├── config/    # Environment configuration
│   │   └── database/  # Database connection
│   └── Dockerfile
├── frontend/          # React frontend
│   ├── src/
│   │   ├── App.js     # Main component
│   │   ├── MyContext.js # Global state management
│   │   ├── components/
│   │   ├── Admin/     # Admin-only features
│   │   └── Search/    # Search functionality
│   └── package.json
└── database/          # Database setup scripts
    ├── *.sql          # Table creation scripts
    ├── *.json         # Dorm room data
    ├── *.py           # Setup scripts
    └── *.ipynb        # Data population notebooks
```

## Troubleshooting

### Docker can't connect to local PostgreSQL
- Use `host.docker.internal` as the SQL_IP instead of `localhost`
- Ensure PostgreSQL is configured to accept connections (check `pg_hba.conf`)

### Authentication issues
- Set `REQUIRE_AUTH="False"` in backend `.env` to bypass auth during development
- For production, ensure Google OAuth is properly configured

### Database connection errors
- Verify PostgreSQL is running: `pg_isready`
- Check credentials in `.env` match your database setup
- For Cloud SQL, ensure your IP is in the authorized networks

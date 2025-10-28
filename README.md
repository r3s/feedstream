# RSS Reader

A minimal RSS feed reader web application built with Go. Everything is vibe coded, except for this sentence you are reading.

## Features

- **Feed Reader** - View feed items with date-based pagination
- **Feed Management** - Add, edit, delete, and organize RSS feeds
- **Import/Export** - Backup and restore feeds as JSON
- **Email and OTP based authentication** - Passwordless login using [Resend](https://resend.com/)

## Environment Variables

**Required:**
- `DATABASE_URL` - PostgreSQL connection string (auto-set by Railway)
- `RESEND_API_KEY` - API key from Resend
- `EMAIL_FROM` - Verified sender email address

**Recommended for Production:**
- `ENVIRONMENT=production` - Enables production mode
- `SESSION_SECRET` - Secret for session encryption (auto-generated if not set)
- `CSRF_SECRET` - Secret for CSRF tokens (auto-generated if not set)

**Optional:**
- `APP_PORT` - Port to run on (default: 8080)

## License

MIT License - see [LICENSE](LICENSE) file for details

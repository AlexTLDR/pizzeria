# 🍕 Pizzeria Website

A modern web application for managing a pizzeria with an admin dashboard, menu management, and secure authentication. Built with Go, Tailwind CSS, and SQLite.

## ✨ Features

- **🔐 Secure Authentication**: Google OAuth integration with email allowlist
- **📋 Menu Management**: Full CRUD operations for menu items
- **👨‍💼 Admin Dashboard**: Intuitive interface for restaurant management
- **📱 Responsive Design**: Mobile-first design using Tailwind CSS
- **🔄 Hot Reload**: Development server with live reload
- **🐳 Docker Support**: Containerized deployment ready
- **⚡ Fast Performance**: Lightweight SQLite database
- **🛡️ Session Security**: Cryptographically signed cookies

## 🛠️ Tech Stack

- **Backend**: Go 1.24+
- **Frontend**: HTML, Tailwind CSS, JavaScript
- **Database**: SQLite
- **Authentication**: Google OAuth 2.0
- **Development**: Air (hot reload), golangci-lint
- **Containerization**: Docker & Docker Compose

## 📋 Prerequisites

Before running this project, ensure you have:

- [Go 1.24+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/) (for Tailwind CSS)
- [Air](https://github.com/air-verse/air) for hot reload (optional)
- [Docker](https://www.docker.com/) (optional, for containerized deployment)

## 🚀 Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/AlexTLDR/pizzeria.git
cd pizzeria
```

### 2. Install Dependencies
```bash
# Install Go dependencies
go mod download

# Install Node.js dependencies
npm install

# Install Air for hot reload (optional)
go install github.com/air-verse/air@latest
```

### 3. Set Up Google OAuth

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API
4. Navigate to **APIs & Services** → **Credentials**
5. Click **Create Credentials** → **OAuth client ID**
6. Select **Web application**
7. Add authorized redirect URI: `http://localhost:8080/auth/google/callback`
8. Note your **Client ID** and **Client Secret**

### 4. Configure Environment Variables

Copy the example environment file and configure it:
```bash
cp .env.example .env
```

Edit `.env` with your values:
```env
GOOGLE_CLIENT_ID=your_google_client_id_here
GOOGLE_CLIENT_SECRET=your_google_client_secret_here
SESSION_SECRET=your_session_secret_here
ALLOWED_EMAILS=admin@example.com,manager@example.com
PORT=8080
```

### 5. Run the Application

#### Development Mode (Recommended)
```bash
# Start with hot reload and CSS watching
make dev
```

#### Alternative Methods
```bash
# Using Air only
make air

# Using Go directly
make run-main

# Build and run
make run
```

Visit `http://localhost:8080` to access the application.

## 📁 Project Structure

```
pizzeria/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── app/            # Application configuration
│   ├── auth/           # Authentication logic
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   └── models/         # Data models
├── static/
│   ├── css/            # Tailwind CSS files
│   ├── js/             # JavaScript files
│   └── images/         # Static images
├── templates/          # HTML templates
├── db/                 # Database files and migrations
├── docs/               # Documentation
├── .air.toml          # Air configuration
├── Dockerfile         # Docker configuration
├── docker-compose.yml # Docker Compose setup
├── Makefile          # Build automation
└── README.md         # This file
```

## 🔧 Available Commands

The project includes a comprehensive Makefile with the following commands:

```bash
# Development
make dev          # Run with hot reload and CSS watching
make air          # Run with Air hot reload only
make run-main     # Run directly with go run

# Building
make build        # Build both Go binary and CSS
make build-go     # Build only Go binary
make build-css    # Build only CSS

# Testing
make test         # Run all tests
make test-v       # Run tests with verbose output
make test-race    # Run tests with race detection
make test-cover   # Run tests with coverage report

# Code Quality
make lint         # Run golangci-lint
make lint-fix     # Run golangci-lint with auto-fix

# Utilities
make clean        # Remove build artifacts
make help         # Show all available commands
```

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)
```bash
docker-compose up -d
```

### Using Docker Manually
```bash
# Build the image
docker build -t pizzeria .

# Run the container
docker run -p 8080:8080 --env-file .env pizzeria
```

## 🔒 Authentication & Security

### Admin Access
1. Navigate to `/login`
2. Sign in with a Google account
3. Only emails listed in `ALLOWED_EMAILS` can access admin features
4. Successful authentication redirects to the admin dashboard

### Security Features
- **OAuth 2.0**: Secure Google authentication
- **Session Management**: Cryptographically signed cookies
- **Email Allowlist**: Restricts admin access to authorized users
- **CSRF Protection**: Built-in protection against cross-site request forgery
- **Secure Headers**: Security headers for enhanced protection

## 🧪 Testing

Run the test suite:
```bash
# Run all tests
make test

# Run with coverage
make test-cover

# Run with race detection
make test-race
```

## 📝 API Endpoints

### Public Routes
- `GET /` - Home page
- `GET /menu` - View menu
- `GET /login` - Login page

### Authentication Routes
- `GET /auth/google/login` - Initiate Google OAuth
- `GET /auth/google/callback` - OAuth callback
- `POST /logout` - Logout

### Admin Routes (Protected)
- `GET /admin` - Admin dashboard
- `GET /admin/menu` - Menu management
- `POST /admin/menu` - Create menu item
- `PUT /admin/menu/:id` - Update menu item
- `DELETE /admin/menu/:id` - Delete menu item

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Guidelines
- Follow Go conventions and best practices
- Write tests for new features
- Run `make lint` before submitting
- Update documentation as needed

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🐛 Issues & Support

Found a bug or have a feature request? Please open an issue on [GitHub](https://github.com/AlexTLDR/pizzeria/issues).

## 🙋‍♂️ Author

**Alex** - [GitHub](https://github.com/AlexTLDR) | [Email](mailto:alex@alextldr.com)

---

**⭐ Star this repository if you found it helpful!**
# Pizzeria Website

A demo website for my local pizzeria with menu management and admin dashboard.

## Admin Authentication

This project uses Google OAuth for admin authentication with secure session management. Only authorized Gmail addresses can access the admin dashboard.

### Security Features

- **Google OAuth**: Uses Google's secure authentication system
- **Email Allowlist**: Only specified Gmail addresses can access admin features
- **Signed Session Cookies**: Sessions are cryptographically signed to prevent tampering
- **Expiration Verification**: Session expiration is verified on each request

### Setup Instructions

1. **Create Google OAuth Credentials**:
   - Go to the [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or use an existing one
   - Navigate to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth client ID"
   - Select "Web application" as the application type
   - Add `http://localhost:8080/auth/google/callback` to the authorized redirect URIs
   - Note your Client ID and Client Secret

2. **Configure Environment Variables**:
   - Copy the `.env.example` file to `.env`
   - Fill in your Google OAuth credentials
   - Add the admin Gmail addresses to the `ALLOWED_EMAILS` variable (comma-separated)

3. **Access the Admin Dashboard**:
   - Start the server
   - Navigate to `http://localhost:8080/login`
   - Sign in with an authorized Gmail account
   - You'll be redirected to the admin dashboard if authorized

## Technologies Used

### Air (Live Reload for Go)
```
go install github.com/air-verse/air@latest
```
https://github.com/air-verse/air/blob/master/README.md

### Tailwind CSS
CSS framework for styling the website.

### SQLite
Lightweight database used to store menu items and flash messages.

### Running the Application
1. Make sure all dependencies are installed
2. Set up the `.env` file with your credentials
3. Run the application:
```
air
```
4. Access the website at `http://localhost:8080`
# Rally Backend API

A Go-based backend API for the Rally application, built with Fiber framework, MongoDB, and Firebase Authentication.

## Features

- **Authentication**: Firebase-based user authentication
- **User Management**: Complete user profile management
- **Database**: MongoDB for data persistence
- **Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Containerization**: Docker support for easy deployment

## Tech Stack

- **Language**: Go 1.24.4
- **Framework**: Gofiber v2
- **Database**: MongoDB
- **Authentication**: Firebase Authentication
- **Documentation**: Swagger/OpenAPI with Swaggo
- **Configuration**: Viper
- **Containerization**: Docker

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register or login user with Firebase token
- `POST /api/v1/auth/login` - Login user with Firebase token

### User Profile Management

- `GET /api/v1/user/me/profile` - Get current user's profile (requires auth)
- `GET /api/v1/user/{id}/profile` - Get user profile by ID
- `PUT /api/v1/user/{id}/profile` - Update user profile (requires auth and ownership)

### Health Check

- `GET /api/v1/health` - Health check endpoint

### Documentation

- `GET /swagger/*` - Swagger UI documentation

## User Profile Fields

The user profile includes the following fields:

- `id` - Unique user identifier
- `email` - User email (from Firebase)
- `displayName` - Display name
- `firstName` - First name
- `lastName` - Last name
- `profilePic` - Profile picture URL
- `bio` - User biography
- `phone` - Phone number
- `dateOfBirth` - Date of birth
- `location` - User location
- `createdAt` - Account creation timestamp
- `updatedAt` - Last update timestamp

## Getting Started

### Prerequisites

- Go 1.24.4 or higher
- MongoDB instance
- Firebase project with service account credentials

### Environment Variables

Create a `.env` file with the following variables:

```env
PORT=8080
ENV=development
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=rally_db
FIREBASE_CREDENTIALS_PATH=serviceAccountKey.json
```

### Installation

1. Clone the repository

```bash
git clone https://github.com/Hoi-Trang-Huynh/rally-backend-api.git
cd rally-backend-api
```

2. Install dependencies

```bash
go mod download
```

3. Set up environment variables (see `.env.example`)

4. Run the application

```bash
make run
# or
go run cmd/server/main.go
```

### Development

- `make run` - Run the application
- `make build` - Build the binary
- `make swag` - Generate Swagger documentation
- `make dev` - Run with hot reload (requires Air)
- `make lint` - Run linter

### Docker

```bash
# Build and run with Docker Compose
docker-compose up --build
```

## Project Structure

```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── handler/        # HTTP handlers
│   ├── infrastructure/ # External services (Firebase, MongoDB)
│   ├── middleware/     # HTTP middleware
│   ├── model/         # Data models and DTOs
│   ├── repository/    # Data access layer
│   ├── router/        # Route definitions
│   └── service/       # Business logic
├── api/docs/          # Generated Swagger documentation
├── bin/               # Build outputs
└── .github/          # CI/CD workflows
```

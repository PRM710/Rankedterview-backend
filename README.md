# RANKEDterview Backend

Go-based backend service for the RANKEDterview platform.

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- MongoDB (local or Atlas)
- Redis (local or cloud)

### Installation

1. **Install dependencies**:
```bash
go mod download
```

2. **Set up environment variables**:
```bash
cp .env.example .env
# Edit .env with your actual values
```

3. **Run the server**:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`.

## 📁 Project Structure

```
backend/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── models/          # Data models
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business logic
│   ├── repositories/    # Data access layer
│   ├── middleware/      # HTTP middleware
│   ├── websocket/       # WebSocket handling
│   ├── signaling/       # WebRTC signaling
│   ├── queue/           # Matchmaking queue
│   ├── database/        # Database connections
│   ├── storage/         # File storage (R2)
│   ├── ai/              # AI integrations
│   ├── oauth/           # OAuth providers
│   └── utils/           # Utility functions
└── pkg/
    └── logger/          # Logging package
```

## 🔌 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/auth/oauth/google` - Google OAuth
- `GET /api/v1/auth/oauth/github` - GitHub OAuth
- `GET /api/v1/auth/callback` - OAuth callback
- `POST /api/v1/auth/refresh` - Refresh token

### Users (Protected)
- `GET /api/v1/users/me` - Get current user
- `PUT /api/v1/users/me` - Update profile
- `GET /api/v1/users/:id` - Get user
- `GET /api/v1/users/:id/stats` - Get statistics

### Matchmaking (Protected)
- `POST /api/v1/matchmaking/join` - Join queue
- `POST /api/v1/matchmaking/leave` - Leave queue
- `GET /api/v1/matchmaking/status` - Queue status

### Rooms (Protected)
- `GET /api/v1/rooms/:roomId` - Get room
- `POST /api/v1/rooms/:roomId/join` - Join room
- `POST /api/v1/rooms/:roomId/leave` - Leave room
- `GET /api/v1/rooms/:roomId/state` - Room state

### Interviews (Protected)
- `GET /api/v1/interviews` - List interviews
- `GET /api/v1/interviews/:id` - Get interview
- `GET /api/v1/interviews/:id/transcript` - Get transcript
- `GET /api/v1/interviews/:id/recording` - Get recording
- `GET /api/v1/interviews/:id/feedback` - Get feedback

### Rankings (Protected)
- `GET /api/v1/rankings/global` - Global leaderboard
- `GET /api/v1/rankings/category/:category` - Category leaderboard
- `GET /api/v1/rankings/user/:userId` - User rank
- `GET /api/v1/rankings/history/:userId` - Rank history

### WebSocket
- `GET /ws` - WebSocket connection

### Health
- `GET /health` - Health check

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services/...
```

## 🔨 Development

### Hot Reload
Install Air for hot reloading:
```bash
go install github.com/cosmtrek/air@latest
air
```

### Linting
```bash
golangci-lint run
```

### Database Migrations
```bash
# TODO: Add migration commands
```

## 🐳 Docker

```bash
# Build image
docker build -t rankedterview-backend .

# Run container
docker run -p 8080:8080 --env-file .env rankedterview-backend
```

## 📝 Environment Variables

See `.env.example` for all required environment variables.

## 🤝 Contributing

Please read the main [CONTRIBUTING.md](../docs/CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - see [LICENSE](../LICENSE) for details.

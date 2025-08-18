# Discord Bot Project

A feature-rich Discord bot built with Go, PostgreSQL, and Docker. This bot manages user birthdays, tracks player statistics from LoL, and provides various community management features.

## Features

- **Birthday Management**: Track and remind users of upcoming birthdays
- **Player Statistics**: Store and retrieve player data with PUUID support
- **Database Integration**: Operations with PostgreSQL and GORM
- **Docker Support**: Containerized application for easy deployment
- **Environment Configuration**: Secure environment variable management

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15
- **ORM**: GORM
- **Containerization**: Docker & Docker Compose
- **Discord Library**: DiscordGo

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Discord Bot Token
- Discord Developer Account

## Installation

### 1. Clone the repository
```bash
git clone https://github.com/machado0/discord-bot-go
cd discord-bot
```

### 2. Create environment file
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```env
# Database Configuration
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=discordbot

# Discord Bot Configuration  
BOT_TOKEN=your_discord_bot_token_here

# Riot API COnfiguration
RIOT_API_
```

### 3. Get Discord Bot Token

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Go to "Bot" section
4. Copy the token and add it to your `.env` file

### 4. Run with Docker

```bash
# Run in background
docker-compose up --build -d

# View logs
docker-compose logs -f discord-bot
```

## 🔨 Development

### Local Development Setup

```bash
# Install dependencies
go mod download

# Run database only
docker-compose up -d postgres

# Run bot locally
go run cmd/discord-bot/main.go
```

## Bot Commands

| Command | Description | Example |
|---------|-------------|---------|
| `!adicionar` | Adds a birthday | `!adicionar 25/12` |
| `!listar` | Lists all birthdays | `!listar` |
| `!remover` | Removes a birthday | `!remover username` |
| `!proximo` | Shows the next upcoming birthdays | `!proximo` |
| `!verificar` | Forces the birthday verification to happen now | `!verificar` |
| `!addcanal` | Sets the messages channel as the channel it will send "Happy Bday" messages to everyone | `!addcanal` |
| `!soloduo` | Shows the player (Hard coded player for now) last ranked game and their League Points | `!soloduo` |
| `!tiltou` | Adds one to the rage counter and shows how many there are and when was the last one | `!tiltou` |
| `!rages` | Shows how many times the person has raged (Hard coded name for now) and when was the last one | `!rages` |
| `!comandos` | Shows all bot commands | `!comandos` |


## Docker Commands

```bash
# Start everything
docker-compose up --build

# Stop everything
docker-compose down

# Rebuild only the bot
docker-compose stop discord-bot
docker-compose up --build -d discord-bot

# View logs
docker-compose logs -f discord-bot
docker-compose logs -f postgres

# Clean restart (⚠ Deletes data)
docker-compose down -v
docker-compose up --build
```

## Database Connection (DBeaver)

Connect to your database using these settings:

- **Host**: `localhost`
- **Port**: `5432`
- **Database**: `[your DB_NAME from .env]`
- **Username**: `[your DB_USER from .env]`
- **Password**: `[your DB_PASSWORD from .env]`

## Project Structure

```
discord-bot/
├── cmd/discord-bot/
│   └── main.go              # Application entry point
├── internal/
│   ├── database/
│   │   └── database.go      # Database connection
│   ├── domain/
│   │   └── modelname.go        # GORM models
│   ├── infra/
│   │   └── riot/        
│   │       └── riot_client.go        # Riot API interaction logic
│   ├── pkg/                 # Features and Command Logic
│   │   └── featurename/
|   |       └── featurename.go      
│   └── util/                # Utils for Riot API
│       ├── HttpError.go
│       ├── Parses.go
│       └── QueueIdentifier.go      
├── docker-compose.yml      # Docker services
├── Dockerfile              # Bot container
├── go.mod                  # Go dependencies
├── go.sum
├── .env.example            # Environment template
├── .gitignore
├── .env
└── README.md
```

## Security

- ✅ Environment variables for secrets
- ✅ `.env` files in `.gitignore`
- ✅ Database password protection
- ✅ Input validation and sanitization
- ✅ Error handling and logging

## Troubleshooting

### Common Issues

**Bot won't start**:
- Check if `BOT_TOKEN` is valid
- Verify Discord bot permissions
- Check Docker logs: `docker-compose logs discord-bot`

**Database connection failed**:
- Verify database is running: `docker-compose ps`
- Check credentials in `.env` file
- Reset database: `docker-compose down -v && docker-compose up --build`

**Permission errors**:
- Bot needs proper Discord permissions
- Check server invite link includes required scopes

## Support

- Create an [Issue](https://github.com/yourusername/discord-bot/issues) for bug reports

## Roadmap

- [ ] Add more birthday reminder features
- [ ] Implement player ranking system
- [ ] Add web dashboard
- [ ] Multi-server support (Currently not working )
- [ ] Backup and restore functionality

---

**Made with ❤️ for my friends by [Ale](https://github.com/yourusername](https://github.com/machado0))**

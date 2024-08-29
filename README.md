<p align="center">
  <img src="https://res.cloudinary.com/diroilukd/image/upload/v1724899914/Designer_4_h5dmas.png" alt="Email Blaze Logo" width="200"/>
</p>

<h1 align="center">Email Blaze</h1>

<p align="center">
  <em>A powerful and secure email sender service built with Go</em>
</p>

<p align="center">
  <a href="#✨-features">Features</a> •
  <a href="#🏗️-project-structure">Project Structure</a> •
  <a href="#⚙️-configuration">Configuration</a> •
  <a href="#🚀-getting-started">Getting Started</a> •
  <a href="#🛣️-api-endpoints">API Endpoints</a> •
  <a href="#🤝-contributing">Contributing</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
</p>

---

## ✨ Features

- ✉️ SMTP email sending
- 🔒 Domain verification (MX and DKIM records)
- 🚦 Rate limiting
- 🔑 JWT-based authentication
- ⚙️ Configurable settings via YAML
- 📝 Logging with Zap

## 🏗️ Project Structure

The project is organized into several packages:

- `pkg/domainverifier`: Handles domain and DKIM record verification
- `internals/logger`: Provides logging functionality using Zap
- `internals/email`: Manages email sending operations
- `internals/ratelimit`: Implements rate limiting for API requests
- `internals/auth`: Handles user authentication and JWT token generation/verification
- `internals/config`: Manages application configuration

## Configuration

Email Blaze uses a YAML configuration file. Create a `config.yaml` file in the project root with the following structure:

```yaml
smtp_port: 25
api_port: 8080
database_url: "postgres://user:password@localhost/emailblaze"
jwt_secret: "your-secret-key"
rate_limit: 100
max_file_size: 10485760 # 10MB in bytes
```

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/email-blaze.git
   cd email-blaze
   ```

2. Create a `config.yaml` file in the project root with your configuration.

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

## API Endpoints (WIP)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/send` | POST | Send an email |
| `/api/verify` | POST | Verify a domain |
| `/api/auth/login` | POST | Authenticate and get JWT token |
| `/api/auth/refresh` | POST | Refresh JWT token |


## Contributing

We welcome contributions to Email Blaze! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

---

<p align="center">
  Made with ❤️ by the Suyash Thakur
</p>
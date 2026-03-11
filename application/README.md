# NC FTTX Portal - Application

**Note**: This application code represents what a DevOps engineer would receive from the development team.

## Application Overview

A simple Go web application for North Carolina FTTX/FTTH permitting information:

- **Business Purpose**: Telecom infrastructure permitting portal
- **Technology**: Go + Bootstrap web interface  
- **Architecture**: Single binary with embedded web assets
- **Data**: North Carolina municipalities with permitting requirements

## Developer Information

- **Language**: Go 1.21+
- **Dependencies**: Minimal (see go.mod)
- **Database**: In-memory data (no external dependencies)
- **Web Framework**: Standard library + simple templating

## Running Locally

```bash
# Install Go dependencies
go mod tidy

# Run application
go run main.go

# Access application
open http://localhost:8080
```

## Endpoints

- `GET /` - Home page with municipality information
- `GET /health` - Health check endpoint
- `GET /metrics` - Basic metrics (for monitoring integration)

## DevOps Notes

This is a **simple, containerizable application** designed to demonstrate:
- CI/CD pipeline implementation
- Kubernetes deployment strategies  
- Monitoring and observability integration
- Production deployment best practices

The application complexity is intentionally minimal to focus on **DevOps engineering skills** rather than application development.

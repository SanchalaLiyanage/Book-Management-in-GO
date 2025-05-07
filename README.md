# Book Management REST API

![Go](https://img.shields.io/badge/Go-1.21+-blue)
![GitHub](https://img.shields.io/badge/license-MIT-green)

A RESTful API for managing books with CRUD operations and advanced search capabilities, built with Go.

## Features

- **CRUD Operations**:
  - Create, Read, Update, and Delete books
  - JSON request/response format
  - Input validation

- **Advanced Search**:
  - Case-insensitive keyword search in titles/descriptions
  - Concurrent search using goroutines
  - Pagination support

- **Infrastructure**:
  - File-based persistence (JSON)
  - Docker containerization
  - Kubernetes deployment ready

## API Endpoints

| Method | Endpoint                | Description                          |
|--------|-------------------------|--------------------------------------|
| GET    | `/books`                | List all books (with pagination)     |
| POST   | `/books`                | Create a new book                    |
| GET    | `/books/{id}`           | Get a specific book                  |
| PUT    | `/books/{id}`           | Update a book                        |
| DELETE | `/books/{id}`           | Delete a book                        |
| GET    | `/books/search?q=term`  | Search books by keyword              |

## Prerequisites

- Go 1.21+
- Docker (optional)
- Kubernetes (optional)
- Postman/curl for testing

## Installation

### 1. Local Development
```bash
# Clone repository
git clone https://github.com/yourusername/book-api.git
cd book-api

# Install dependencies
go mod download

# Run the application
go run main.go
```

### 2. Using Docker
```bash
# Build image
docker build -t book-api .

# Run container
docker run -p 8080:8080 book-api
```

### 3. Kubernetes Deployment
```bash
# Start Minikube
minikube start

# Build and deploy
kubectl apply -f k8s/

# Access service
minikube service book-api
```

## Testing

### Unit Tests
```bash
go test ./... -v
```

### API Tests (Postman)
Run in Postman

#### Sample Requests:
##### Create Book:
```bash
curl -X POST http://localhost:8080/books \
-H "Content-Type: application/json" \
-d '{
  "title": "Sample Book",
  "authorId": "e0d91f68-a183-477d-8aa4-1f44ccc78a70",
  "isbn": "9781234567890",
  "pages": 300
}'
```

##### Search Books:
```bash
curl "http://localhost:8080/books/search?q=gatsby&limit=5"
```

## Project Structure
```
book-api/
├── data/               # JSON data storage
├── handlers/           # HTTP handlers
├── models/             # Data models
├── repository/         # Data persistence layer
├── k8s/                # Kubernetes manifests
├── main.go             # Application entry point
├── Dockerfile          # Container configuration
└── go.mod              # Dependency management
```

## Configuration

### Environment Variables:
```env
PORT=8080               # Server port
DATA_FILE=./data/books.json  # Data storage path
```

## Technical Highlights

- **Concurrent Search**: Uses goroutines and channels for parallel processing
- **Atomic Writes**: Safe file operations with mutex locks
- **Validation**: Comprehensive input validation
- **Error Handling**: Custom error types with proper HTTP status codes

## License

MIT License - See LICENSE for details.

Developed by: Navodi Sanchala Liyanage
For: Software Engineering Intern Assignment  
Date: 29th March, 2025


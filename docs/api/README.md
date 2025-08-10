# ProveMySelf API Documentation

Welcome to the ProveMySelf API documentation. This API powers the AI-powered quiz platform for creating and managing interactive assessments.

## Table of Contents

- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [API Reference](#api-reference)
- [Examples](#examples)
- [SDKs and Tools](#sdks-and-tools)

## Quick Start

### Base URL

- **Development**: `http://localhost:8080`
- **Production**: `https://api.provemyself.com`

### Making Your First Request

```bash
# Health check
curl http://localhost:8080/health

# List projects
curl http://localhost:8080/api/v1/projects
```

### Response Format

All API responses follow a consistent JSON format:

```json
{
  "data": {
    // Response data here
  }
}
```

Error responses follow this format:

```json
{
  "error": {
    "code": "error_code",
    "message": "Human-readable error message",
    "details": "Additional error details (optional)"
  }
}
```

## Authentication

The API uses JWT (JSON Web Token) authentication. Include your token in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/v1/projects
```

### Getting a Token

Authentication endpoints will be available in Stage 1. For Stage 0 development, authentication is optional.

## Rate Limiting

API requests are rate-limited to ensure fair usage:

- **Development**: 1000 requests per minute per IP
- **Production**: 100 requests per minute per authenticated user

Rate limit headers are included in all responses:

- `X-RateLimit-Limit`: Request limit per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Timestamp when the limit resets

When the rate limit is exceeded, the API returns a `429 Too Many Requests` response.

## Error Handling

### HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `204 No Content` - Request successful, no response body
- `400 Bad Request` - Invalid request format
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., already exists)
- `422 Unprocessable Entity` - Validation error
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### Error Response Format

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Request validation failed",
    "errors": [
      {
        "field": "title",
        "tag": "required",
        "message": "title is required"
      }
    ]
  }
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `invalid_request_body` | Request body is malformed JSON |
| `validation_failed` | One or more fields failed validation |
| `project_not_found` | Project with given ID doesn't exist |
| `project_already_published` | Attempt to publish already published project |
| `unauthorized` | Authentication token missing or invalid |
| `forbidden` | Insufficient permissions for requested operation |
| `rate_limited` | Too many requests, slow down |
| `internal_error` | Unexpected server error |

## API Reference

### System Endpoints

#### Health Check
```
GET /health
```

Returns the overall health of the API and its dependencies.

**Response Example:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "services": {
    "database": "healthy",
    "storage": "healthy"
  }
}
```

#### Liveness Probe
```
GET /health/live
```

Simple liveness check for container orchestration.

#### Readiness Probe
```
GET /health/ready
```

Readiness check that validates all dependencies.

#### System Metrics
```
GET /metrics
```

Returns detailed system metrics including memory usage, garbage collection stats, and performance data.

### Project Endpoints

#### List Projects
```
GET /api/v1/projects
```

**Query Parameters:**
- `limit` (optional): Maximum number of projects (1-100, default: 20)
- `offset` (optional): Number of projects to skip (default: 0)
- `search` (optional): Search term for title/description
- `tags` (optional): Comma-separated list of tags to filter by

**Response Example:**
```json
{
  "projects": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "title": "JavaScript Basics Quiz",
      "description": "Test your knowledge of JavaScript fundamentals",
      "tags": ["javascript", "beginner"],
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "published_at": "2024-01-15T11:00:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

#### Create Project
```
POST /api/v1/projects
```

**Request Body:**
```json
{
  "title": "My New Quiz",
  "description": "A comprehensive quiz about web development",
  "tags": ["web", "html", "css"]
}
```

**Validation Rules:**
- `title`: Required, 1-200 characters
- `description`: Optional, max 1000 characters
- `tags`: Optional, max 10 tags, each max 50 characters

#### Get Project
```
GET /api/v1/projects/{projectId}
```

Returns details for a specific project.

#### Update Project
```
PUT /api/v1/projects/{projectId}
```

**Request Body:**
```json
{
  "title": "Updated Quiz Title",
  "description": "Updated description",
  "tags": ["updated", "tags"]
}
```

#### Delete Project
```
DELETE /api/v1/projects/{projectId}
```

Permanently deletes a project. This action cannot be undone.

#### Publish Project
```
POST /api/v1/projects/{projectId}/publish
```

Marks a project as published. Once published, a project cannot be unpublished.

## Examples

### Creating a Project

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "title": "React Fundamentals Quiz",
    "description": "Test your understanding of React concepts",
    "tags": ["react", "javascript", "frontend"]
  }'
```

### Searching Projects

```bash
# Search by title/description
curl "http://localhost:8080/api/v1/projects?search=javascript"

# Filter by tags
curl "http://localhost:8080/api/v1/projects?tags=beginner,javascript"

# Pagination
curl "http://localhost:8080/api/v1/projects?limit=10&offset=20"
```

### Publishing a Project

```bash
curl -X POST http://localhost:8080/api/v1/projects/123e4567-e89b-12d3-a456-426614174000/publish \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## SDKs and Tools

### OpenAPI Specification

The complete API specification is available in OpenAPI 3.0 format:
- [openapi.yaml](./openapi.yaml)

### Code Generation

Generate client SDKs using the OpenAPI specification:

```bash
# Generate TypeScript client
npm install @openapitools/openapi-generator-cli
openapi-generator-cli generate -i openapi.yaml -g typescript-fetch -o ./client-typescript

# Generate Python client
openapi-generator-cli generate -i openapi.yaml -g python -o ./client-python

# Generate Go client
openapi-generator-cli generate -i openapi.yaml -g go -o ./client-go
```

### Postman Collection

Import the API into Postman using the OpenAPI specification for interactive testing.

### curl Examples

Complete curl examples are available in [examples/](./examples/) directory.

## Support

- **Issues**: Report bugs and feature requests on [GitHub](https://github.com/provemyself/api/issues)
- **Email**: Technical support at api-support@provemyself.com
- **Documentation**: Latest docs at [docs.provemyself.com](https://docs.provemyself.com)

## Changelog

### Version 1.0.0 (Stage 0)
- Initial API release
- Project CRUD operations
- Health monitoring endpoints
- System metrics
- Comprehensive error handling
- Rate limiting
- OpenAPI 3.0 specification
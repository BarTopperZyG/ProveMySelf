# ProveMySelf API Guide

This guide provides comprehensive documentation for the ProveMySelf API, covering authentication, endpoints, error handling, and best practices.

## Table of Contents
1. [Authentication](#authentication)
2. [Rate Limiting](#rate-limiting)
3. [Error Handling](#error-handling)
4. [Pagination](#pagination)
5. [API Endpoints](#api-endpoints)
6. [Examples](#examples)
7. [SDKs and Tools](#sdks-and-tools)

## Authentication

### JWT Bearer Tokens

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Optional Authentication

Some endpoints support optional authentication:
- **With authentication**: Returns user-specific data
- **Without authentication**: Returns public data

### Token Expiration

Tokens expire after 24 hours. Use the refresh endpoint to obtain new tokens without re-authentication.

## Rate Limiting

- **Default limit**: 100 requests per minute per IP address
- **Headers returned**:
  - `X-RateLimit-Limit`: Request limit per window
  - `X-RateLimit-Remaining`: Requests remaining in current window
  - `X-RateLimit-Reset`: Unix timestamp when the limit resets

### Rate Limit Response

```json
{
  "error": {
    "code": "rate_limited",
    "message": "Rate limit exceeded. Please try again later."
  }
}
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "error_code",
    "message": "Human-readable error message",
    "details": "Optional additional details"
  }
}
```

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `internal_error` | 500 | Unexpected server error |
| `validation_failed` | 400 | Request data validation failed |
| `unauthorized` | 401 | Authentication required |
| `forbidden` | 403 | Insufficient permissions |
| `not_found` | 404 | Resource not found |
| `rate_limited` | 429 | Rate limit exceeded |
| `project_not_found` | 404 | Specific project not found |
| `file_too_big` | 413 | File exceeds size limit |
| `invalid_file_type` | 415 | File type not allowed |

## Pagination

List endpoints support pagination with these parameters:

- `limit`: Maximum items to return (1-100, default 20)
- `offset`: Number of items to skip (default 0)

### Pagination Response

```json
{
  "projects": [...],
  "total": 150,
  "limit": 20,
  "offset": 40
}
```

### Navigation

- **First page**: `offset=0`
- **Next page**: `offset = current_offset + limit`
- **Has more**: `offset + limit < total`

## API Endpoints

### System

#### GET /api/v1/health

Health check endpoint for monitoring.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "services": {
    "database": "healthy",
    "storage": "healthy"
  }
}
```

### Projects

#### GET /api/v1/projects

List projects with optional filtering.

**Parameters:**
- `limit` (optional): Items per page (1-100, default 20)
- `offset` (optional): Items to skip (default 0)
- `search` (optional): Search term for title/description
- `tags` (optional): Comma-separated tags to filter

**Example:**
```bash
curl "https://api.provemyself.com/v1/projects?limit=10&search=javascript&tags=beginner"
```

#### POST /api/v1/projects

Create a new project. **Requires authentication.**

**Request Body:**
```json
{
  "title": "My Quiz Project",
  "description": "A comprehensive quiz about web development",
  "tags": ["web", "javascript", "beginner"]
}
```

**Validation Rules:**
- `title`: Required, 1-200 characters
- `description`: Optional, max 1000 characters  
- `tags`: Optional array, max 10 tags, each max 50 characters

#### GET /api/v1/projects/{projectId}

Get a specific project by ID.

**Path Parameters:**
- `projectId`: UUID of the project

**Example:**
```bash
curl "https://api.provemyself.com/v1/projects/123e4567-e89b-12d3-a456-426614174000"
```

## Examples

### Creating a Project

```javascript
const response = await fetch('https://api.provemyself.com/v1/projects', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer your-jwt-token',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    title: 'JavaScript Fundamentals',
    description: 'Test your knowledge of JavaScript basics',
    tags: ['javascript', 'programming', 'beginner']
  })
});

const project = await response.json();
console.log('Created project:', project);
```

### Listing Projects with Pagination

```javascript
async function getAllProjects() {
  const allProjects = [];
  let offset = 0;
  const limit = 50;

  while (true) {
    const response = await fetch(
      `https://api.provemyself.com/v1/projects?limit=${limit}&offset=${offset}`
    );
    const data = await response.json();
    
    allProjects.push(...data.projects);
    
    // Check if we have more pages
    if (offset + limit >= data.total) {
      break;
    }
    
    offset += limit;
  }

  return allProjects;
}
```

### Error Handling

```javascript
async function createProject(projectData) {
  try {
    const response = await fetch('https://api.provemyself.com/v1/projects', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(projectData)
    });

    if (!response.ok) {
      const error = await response.json();
      
      switch (error.error.code) {
        case 'validation_failed':
          console.error('Validation error:', error.error.details);
          break;
        case 'unauthorized':
          console.error('Please log in to create projects');
          break;
        case 'rate_limited':
          console.error('Too many requests, please wait');
          break;
        default:
          console.error('Unexpected error:', error.error.message);
      }
      
      throw new Error(error.error.message);
    }

    return await response.json();
  } catch (error) {
    console.error('Network error:', error);
    throw error;
  }
}
```

## SDKs and Tools

### TypeScript/JavaScript Client

```bash
npm install @provemyself/api-client
```

```javascript
import { createApiClient } from '@provemyself/api-client';

const client = createApiClient('https://api.provemyself.com/v1', {
  headers: { 'Authorization': 'Bearer your-token' }
});

// Usage
const projects = await client.projects.list({ limit: 10 });
const project = await client.projects.create({
  title: 'My Quiz',
  description: 'A great quiz'
});
```

### OpenAPI Tools

- **OpenAPI Spec**: Available at `/openapi.yaml`
- **Swagger UI**: Available at `/docs`
- **Postman Collection**: Import the OpenAPI spec

### Development Tools

```bash
# Validate API responses
curl -X GET "https://api.provemyself.com/v1/health" | jq

# Pretty print JSON responses
curl -s "https://api.provemyself.com/v1/projects" | jq '.projects[] | {id, title}'

# Test with different content types
curl -X POST "https://api.provemyself.com/v1/projects" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token" \
  -d '{"title": "Test Project"}'
```

## Best Practices

### Request Optimization

1. **Use appropriate page sizes**: Balance between API calls and memory usage
2. **Implement caching**: Cache responses when appropriate
3. **Use conditional requests**: Leverage ETags when available
4. **Batch operations**: Group related operations when possible

### Error Handling

1. **Always check response status**: Don't assume success
2. **Handle specific error codes**: Provide meaningful user feedback
3. **Implement retry logic**: For transient errors (500s, rate limits)
4. **Log errors appropriately**: Include request IDs for debugging

### Security

1. **Store tokens securely**: Never expose JWT tokens in client-side code
2. **Use HTTPS**: Always use secure connections
3. **Validate input**: Even though the API validates, client validation improves UX
4. **Handle token expiration**: Implement automatic token refresh

### Performance

1. **Use appropriate HTTP methods**: GET for reading, POST for creating, etc.
2. **Minimize payload size**: Only send necessary data
3. **Use compression**: Enable gzip compression
4. **Monitor rate limits**: Track usage and implement backoff strategies
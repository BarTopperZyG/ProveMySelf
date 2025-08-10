# ProveMySelf Architecture Documentation

This document provides a comprehensive overview of the ProveMySelf platform architecture, design decisions, and implementation patterns.

## Table of Contents
1. [System Overview](#system-overview)
2. [Core Architecture](#core-architecture)
3. [Technology Stack](#technology-stack)
4. [Domain Model](#domain-model)
5. [API Design](#api-design)
6. [Security Architecture](#security-architecture)
7. [Data Flow](#data-flow)
8. [Deployment Architecture](#deployment-architecture)
9. [Scalability Considerations](#scalability-considerations)
10. [Design Decisions](#design-decisions)

## System Overview

ProveMySelf is an AI-powered quiz creation platform designed as a "Canva for quizzes." The platform enables users to create interactive, accessible quizzes using a visual drag-and-drop interface.

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Quiz Studio   │    │   Quiz Player   │    │   Admin Panel   │
│   (Frontend)    │    │   (Frontend)    │    │   (Frontend)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                        │                        │
         └────────────────────────┼────────────────────────┘
                                  │
                    ┌─────────────────┐
                    │   API Gateway   │
                    │   (Next.js)     │
                    └─────────────────┘
                                  │
                    ┌─────────────────┐
                    │  Backend API    │
                    │    (Go)         │
                    └─────────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         │                       │                        │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │ Object Storage  │    │   Redis Cache   │
│   (Database)    │    │   (Files)       │    │   (Sessions)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Architecture

### Monorepo Structure

The project follows a monorepo pattern with clear separation of concerns:

```
ProveMySelf/
├── backend/go/              # Go API backend
│   ├── cmd/api/            # Application entry point
│   ├── internal/           # Internal packages
│   │   ├── auth/          # Authentication logic
│   │   ├── config/        # Configuration management
│   │   ├── core/          # Business logic (domain)
│   │   ├── http/          # HTTP handlers and middleware
│   │   ├── store/         # Data access layer
│   │   └── types/         # DTOs and response types
│   ├── pkg/               # Reusable packages
│   └── test/              # Integration tests
├── frontend/              # Frontend applications
│   ├── studio/            # Quiz creation interface
│   └── player/            # Quiz consumption interface
├── packages/              # Shared packages
│   ├── openapi/           # API specification
│   ├── openapi-client/    # Generated TypeScript client
│   ├── schemas/           # Zod validation schemas
│   └── ui-tokens/         # Design system tokens
└── docs/                  # Documentation
```

### Layered Architecture (Backend)

The backend follows a clean architecture pattern:

#### 1. HTTP Layer (`internal/http/`)
- **Handlers**: Process HTTP requests/responses
- **Middleware**: Cross-cutting concerns (auth, logging, validation)
- **Routes**: URL routing configuration

#### 2. Core/Domain Layer (`internal/core/`)
- **Services**: Business logic implementation
- **Entities**: Domain models
- **Interfaces**: Abstractions for external dependencies

#### 3. Store Layer (`internal/store/`)
- **Repositories**: Data access implementations
- **Models**: Database-specific models
- **Migrations**: Database schema changes

#### 4. Infrastructure Layer
- **Config**: Application configuration
- **Auth**: Authentication/authorization
- **Storage**: File storage abstractions

## Technology Stack

### Backend
- **Language**: Go 1.22+
- **Web Framework**: Chi (lightweight, composable)
- **Database**: PostgreSQL 15+
- **Validation**: go-playground/validator
- **Logging**: zerolog (structured logging)
- **Testing**: testify + testcontainers

### Frontend
- **Framework**: Next.js 14+ (App Router)
- **Language**: TypeScript (strict mode)
- **Styling**: Tailwind CSS + shadcn/ui
- **State Management**: Zustand
- **Forms**: React Hook Form + Zod
- **Testing**: Vitest + React Testing Library

### Shared/Infrastructure
- **API Spec**: OpenAPI 3.0
- **Package Manager**: pnpm (workspaces)
- **Build Tool**: Native (Go) + Next.js
- **CI/CD**: GitHub Actions
- **Containerization**: Docker

## Domain Model

### Core Entities

#### Project
```go
type Project struct {
    ID          string     `json:"id"`
    Title       string     `json:"title"`
    Description *string    `json:"description"`
    Tags        []string   `json:"tags"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    PublishedAt *time.Time `json:"published_at"`
}
```

#### Quiz Item
```go
type QuizItem struct {
    ID         string                 `json:"id"`
    ProjectID  string                 `json:"project_id"`
    Type       QuizItemType          `json:"type"`
    Title      string                 `json:"title"`
    Content    map[string]interface{} `json:"content"`
    Position   int                    `json:"position"`
    CreatedAt  time.Time             `json:"created_at"`
    UpdatedAt  time.Time             `json:"updated_at"`
}
```

#### Quiz Bundle (Export Format)
- **ui.json**: Adaptive Cards UI definition
- **quiz.json**: QTI 3.0 assessment structure
- **assets/**: Media files and resources

### Standards Integration

#### Adaptive Cards (UI Layer)
```json
{
  "type": "AdaptiveCard",
  "version": "1.5",
  "body": [
    {
      "type": "TextBlock",
      "text": "Which programming language is best for beginners?",
      "wrap": true
    },
    {
      "type": "Input.ChoiceSet",
      "id": "response",
      "choices": [
        {"title": "Python", "value": "python"},
        {"title": "JavaScript", "value": "javascript"}
      ],
      "data": {"itemId": "item-1", "responseKey": "choice"}
    }
  ]
}
```

#### QTI 3.0 (Assessment Layer)
```xml
<qti-assessment-item identifier="item-1" title="Programming Languages">
  <response-declaration identifier="RESPONSE" cardinality="single" base-type="identifier">
    <correct-response>
      <value>python</value>
    </correct-response>
  </response-declaration>
  <item-body>
    <choice-interaction response-identifier="RESPONSE" shuffle="false" max-choices="1">
      <prompt>Which programming language is best for beginners?</prompt>
      <simple-choice identifier="python">Python</simple-choice>
      <simple-choice identifier="javascript">JavaScript</simple-choice>
    </choice-interaction>
  </item-body>
</qti-assessment-item>
```

## API Design

### RESTful Principles

- **Resources**: Nouns (projects, items, attempts)
- **HTTP Methods**: Semantic usage (GET, POST, PUT, DELETE)
- **Status Codes**: Appropriate HTTP status codes
- **Stateless**: Each request contains all necessary information

### URL Structure
```
/api/v1/health
/api/v1/projects
/api/v1/projects/{id}
/api/v1/projects/{id}/items
/api/v1/projects/{id}/items/{itemId}
/api/v1/projects/{id}/publish
/api/v1/attempts
/api/v1/attempts/{id}
```

### Error Handling
Consistent error envelope across all endpoints:
```json
{
  "error": {
    "code": "validation_failed",
    "message": "Validation failed",
    "details": "Field 'title' is required but was not provided"
  }
}
```

### Response Patterns

#### Success Response
```json
{
  "id": "123",
  "title": "My Quiz",
  "created_at": "2024-01-01T12:00:00Z"
}
```

#### List Response (with Pagination)
```json
{
  "projects": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

## Security Architecture

### Authentication Flow
1. User provides credentials
2. Server validates and issues JWT token
3. Client includes token in Authorization header
4. Server validates token on each request
5. Token expires after configurable time

### Authorization Layers
- **Route-level**: Authentication required/optional
- **Resource-level**: Owner-based access control
- **Action-level**: Role-based permissions

### Security Middleware Stack
```go
r.Use(middleware.SecurityHeaders)    // Security headers
r.Use(middleware.RateLimit)          // Rate limiting
r.Use(middleware.RequestSizeLimit)   // Request size limits
r.Use(middleware.CORS)               // CORS configuration
r.Use(middleware.AuthenticateJWT)    // JWT validation
```

### Security Headers
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `X-Frame-Options: DENY`
- `Strict-Transport-Security`
- Content Security Policy

## Data Flow

### Quiz Creation Flow
```
Studio UI → Canvas Editor → Property Inspector → Export Pipeline → Quiz Bundle
    ↓
Backend API → Project Service → Database → File Storage
    ↓
Player Runtime → Adaptive Cards Renderer → Response Collection → Scoring
```

### Response Collection Flow
```
Player UI → Form Submission → Response Validation → Backend API
    ↓
Attempt Service → Response Processing → Score Calculation
    ↓
xAPI Events → LRS → Analytics Dashboard
```

### File Upload Flow
```
Frontend → File Validation → Upload Endpoint → Storage Service
    ↓
Local/S3 Storage → Metadata Database → Public URL Generation
```

## Deployment Architecture

### Development Environment
```
Developer Machine:
├── Backend (localhost:8080)
├── Studio (localhost:3000)
├── Player (localhost:3001)
└── PostgreSQL (localhost:5432)
```

### Production Environment
```
Load Balancer → API Gateway → Backend Instances
                    ↓
            Database Cluster ← Read Replicas
                    ↓
              Object Storage (CDN)
                    ↓
                Redis Cluster
```

### Container Strategy
- **Backend**: Multi-stage Go build
- **Frontend**: Next.js optimized build
- **Database**: Managed PostgreSQL service
- **Storage**: Object storage service (S3-compatible)

## Scalability Considerations

### Horizontal Scaling
- **Stateless Backend**: Multiple API instances behind load balancer
- **Database**: Read replicas for query scaling
- **File Storage**: CDN for global asset distribution
- **Caching**: Redis for session and response caching

### Performance Optimizations
- **Connection Pooling**: Database connection management
- **Query Optimization**: Indexed queries, pagination
- **Asset Optimization**: Compressed images, lazy loading
- **API Caching**: Response caching for read-heavy endpoints

### Monitoring and Observability
- **Structured Logging**: Request tracing with correlation IDs
- **Metrics Collection**: Performance and business metrics
- **Health Checks**: Service availability monitoring
- **Error Tracking**: Centralized error reporting

## Design Decisions

### Technology Choices

#### Why Go for Backend?
- **Performance**: Fast compilation and execution
- **Simplicity**: Easy to read and maintain
- **Concurrency**: Built-in goroutines for handling concurrent requests
- **Standard Library**: Rich standard library reduces dependencies
- **Deployment**: Single binary deployment

#### Why Next.js for Frontend?
- **Full-Stack**: API routes for backend functionality
- **Performance**: Server-side rendering and optimization
- **Developer Experience**: Hot reloading, TypeScript support
- **Ecosystem**: Rich component library ecosystem
- **Deployment**: Vercel optimization and edge functions

#### Why PostgreSQL?
- **ACID Compliance**: Data consistency and reliability
- **JSON Support**: Flexible schema for quiz content
- **Extensions**: Full-text search, UUID generation
- **Scaling**: Mature replication and partitioning
- **Standards**: SQL standard compliance

### Architectural Patterns

#### Monorepo vs Polyrepo
**Chosen**: Monorepo
- **Benefits**: Shared code, atomic changes, single CI/CD
- **Trade-offs**: Repository size, build complexity

#### Microservices vs Monolith
**Chosen**: Modular Monolith
- **Benefits**: Simpler deployment, easier development
- **Future**: Can split into microservices as needed

#### Event-Driven vs Request-Response
**Chosen**: Hybrid approach
- **Synchronous**: CRUD operations
- **Asynchronous**: Analytics, notifications (future)

### Standards Adoption

#### Why Adaptive Cards?
- **Portability**: Cross-platform UI rendering
- **Standardization**: Microsoft-backed standard
- **Flexibility**: Extensible action system
- **Tooling**: Existing designer tools and libraries

#### Why QTI 3.0?
- **Interoperability**: LMS integration capability
- **Assessment Features**: Rich scoring and feedback
- **Industry Standard**: Widely adopted in education
- **Future-Proof**: Latest version with modern features

#### Why xAPI?
- **Analytics**: Detailed learning analytics
- **Standards Compliance**: ADL specification
- **Flexibility**: Custom statement types
- **Integration**: LRS ecosystem compatibility
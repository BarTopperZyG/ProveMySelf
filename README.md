# ProveMySelf 🛠️

![Awesome Badge](https://awesome.re/badge.svg)

---

## 📌 Logo
> *(Placeholder for your logo — replace with `docs/logo.png` once ready)*

---

## 📝 Short Description
**ProveMySelf** is an interactive quiz and assessment platform — think *Canva for quizzes*.  
It enables creators to design beautiful, accessible, and interactive quizzes with multiple question types, then publish them anywhere.  

Built as a modern **monorepo**, it combines a **Go 1.22+ backend** with a **Next.js 14+ frontend**, sharing types, schemas, and generated API clients.

---

## 💡 Why `.cursorrules`?
The `.cursorrules` file in this repo defines **strict AI-assisted coding guidelines**:
- Enforces project structure and naming conventions
- Guarantees tests, documentation, and accessibility
- Keeps backend and frontend aligned on contracts via OpenAPI & Zod
- Helps AI tools like Cursor start building in the right place, with the right stack

---

## 📚 Table of Contents
1. [About the App](#about-the-app)
2. [Tech Stack](#tech-stack)
3. [Monorepo Structure](#monorepo-structure)
4. [Development Setup](#development-setup)
5. [How to Start Building](#how-to-start-building)
6. [Key Rules Summary](#key-rules-summary)
7. [Contributing](#contributing)
8. [License](#license)

---

## 📖 About the App
**ProveMySelf** is designed to:
- Let users build **interactive quizzes** with drag-and-drop ease
- Support multiple question types (MCQ, drag-drop, hotspots, ordering, fill-in-the-blank, etc.)
- Offer instant preview, themes, and brand kits
- Track analytics via **xAPI**
- Integrate with LMSes via **LTI 1.3**
- Export/import in **QTI 3.0** for portability
- Work **mobile-first** and meet **WCAG 2.2 AA** accessibility standards

---

## 🛠 Tech Stack

### Backend
- **Go 1.22+** with Chi router
- **PostgreSQL** database
- **go-playground/validator** for validation
- **zerolog/slog** for structured logging
- **OpenAPI** generation from Go comments
- **Testcontainers** for integration testing

### Frontend
- **Next.js 14+** (App Router) with React 18
- **TypeScript strict** mode
- **Tailwind CSS** + **shadcn/ui** components
- **Zustand** for local state management
- **React Hook Form** + **Zod** for form handling
- **Vitest** + **React Testing Library** for testing
- **axe-core** for automated accessibility checks

### Shared
- **npm workspaces** for shared packages
- **Zod schemas** in `packages/schemas`
- **Generated OpenAPI client** in `packages/openapi-client`

---

## 🗂 Monorepo Structure
```
ProveMySelf/
├── backend/go/          # Go backend API
├── frontend/next/       # Next.js frontend
├── packages/
│   ├── schemas/         # Zod schemas
│   ├── ui-tokens/       # Design tokens
│   └── openapi-client/  # Generated API client
├── docs/                # Documentation
├── .github/workflows/   # CI/CD workflows
├── Makefile             # Monorepo build orchestrator
└── package.json         # npm workspaces root
```

---

## 🚀 Development Setup

### Prerequisites
- **Go 1.22+**
- **Node.js 20+**
- **pnpm ≥ 8**
- **PostgreSQL 15+**
- **Docker** (for integration tests)

### Quick Start

#### 1. Clone and Install
```bash
git clone <your-repo-url>
cd ProveMySelf
pnpm install
```

#### 2. Backend Setup
```bash
cd backend/go
go mod tidy
make dev
```

#### 3. Frontend Setup
```bash
# From root directory
pnpm dev --filter frontend/next
```

#### 4. Run All Services
```bash
# From root directory
make dev
```

---

## 🏗 How to Start Building

If you are **Cursor AI** (or a developer following `.cursorrules`):

### Choose the Correct Folder Based on Feature
- **Backend API** → `backend/go/internal/...`
- **Frontend UI** → `frontend/next/...`
- **Shared schemas or clients** → `packages/...`

### Follow Naming Conventions
- **camelCase** for variables and functions
- **PascalCase** for types and interfaces
- **kebab-case** for files and directories

### Adding an API Endpoint
1. Create request/response DTO in `internal/types`
2. Add handler in `internal/http`
3. Implement logic in `internal/core`
4. Add tests (`*_test.go`)
5. Document in OpenAPI comments

### Adding a UI Component
1. Place in `frontend/next/components`
2. Export typed props interface
3. Add JSDoc with usage example
4. Test with RTL/Vitest
5. Ensure WCAG compliance

### Updating Documentation
- Update README if public API changes
- Add feature docs to `/docs`

---

## 📋 Key Rules Summary

### ✅ Always Do
- **Never hardcode secrets** — always use environment variables
- **Always include tests and docs** with new features
- **Keep backend and frontend contracts in sync** via OpenAPI
- **Maintain ≥70% coverage** in backend core packages
- **Accessibility is non-negotiable** for interactive UI

### ❌ Never Do
- Ship code without tests
- Hardcode configuration values
- Skip documentation updates
- Ignore accessibility requirements

---

## 🤝 Contributing

### Getting Started
1. **Fork and clone** the repository
2. **Install dependencies**: `pnpm install`
3. **Follow `.cursorrules`** for file placement and standards

### Pull Request Requirements
- ✅ **Tests** included
- ✅ **Documentation** updated
- ✅ **Lint/type checks** passing
- ✅ **Accessibility** verified (for UI changes)

---

## 📄 License
*[Add your license information here]*

---

## 🔗 Quick Links
- [`.cursorrules`](./.cursorrules) - AI coding guidelines
- [`docs/`](./docs/) - Detailed documentation
- [`Makefile`](./Makefile) - Build commands
- [Issues](../../issues) - Report bugs or request features



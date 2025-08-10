# ProveMySelf — Working Plan (Roadmap)

## 📋 Table of Contents
1. [Goal & Philosophy](#goal--philosophy)
2. [Core Decision: Declarative Formats](#core-decision-declarative-formats)
3. [High-Level Architecture](#high-level-architecture)
4. [Deliverables by Phase](#deliverables-by-phase)
5. [Development Process](#development-process)
6. [Declarative Artifacts](#declarative-artifacts)
7. [Binding Model](#binding-model)
8. [Acceptance Criteria](#acceptance-criteria)
9. [Risks & Mitigations](#risks--mitigations)
10. [Next Tasks](#next-tasks)
11. [Example Snippets](#example-snippets)
12. [Definitions & Links](#definitions--links)
13. [Definition of Done](#definition-of-done)

---

## 🎯 Goal & Philosophy

Build an **AI-powered, Canva-like quiz studio** where creators design interactive assessments visually.

### Core Principles
- **The UI is the product.** The "studio" emits a **declarative artifact** that fully describes layout, interactions, logic, and semantics.
- **Backend is headless**: storage, versioning, scoring, analytics, and integrations. It cooperates with whatever the studio emits, via stable contracts (OpenAPI + events).
- **Standards over invention**: use proven, widely supported formats so authored content is portable and future-safe.

---

## 🔧 Core Decision: Declarative Formats

To let users "make the UI" and keep the backend generic, we compose two battle-tested specs:

### 1. **UI Layout & Interactions → Adaptive Cards JSON**
- Portable, declarative UI snippets in JSON; host app maps them to native components.
- Mature ecosystem, multiple renderers, versioned schema with fallback.

### 2. **Assessment Semantics → IMS QTI 3.0 (items/tests/results)**
- Industry standard for questions, tests, outcomes. Ensures content can round-trip with LMS/item banks.

### 3. **Analytics/Telemetry → xAPI (Experience API) to an LRS**
- Emit "actor–verb–object–result" statements; store in any LRS.

> **Alternative (for form-heavy authoring UIs inside the studio)**: **JSON Forms / react-jsonschema-form `uiSchema`** to quickly scaffold property panels and editors. These are **studio-internal** only (not the public quiz artifact).

### Why Not Invent a New Language Now?
- Adaptive Cards (+ custom host actions) covers our visual layer
- QTI covers quiz semantics
- xAPI covers analytics
- New DSLs add friction and maintenance without clear benefit at MVP.

---

## 🏗 High-Level Architecture

### Studio (Next.js 14, React, TS)
- Drag-and-drop canvas, property inspector, templates, live preview.
- Produces **Quiz Bundle** = `quiz.json` (QTI 3.0 items/tests) + `ui.json` (Adaptive Cards) + assets.
- Real-time co-editing via **Yjs** provider (presence, undo/redo, offline merge).

### Player (Next.js Edge / lightweight runtime)
- Renders **Adaptive Cards** into React components; binds controls to QTI item logic (responses, scoring hooks).
- Works offline (PWA) with asset caching.

### Backend (Go 1.22, Chi)
- Headless APIs: projects, items, attempts, scoring, publish, webhooks.
- Stores quiz bundles (Postgres + object storage).
- Scores via QTI-compatible rules; emits **xAPI** events to an LRS.

### Contracts
- **OpenAPI**: request/response models for all services.
- **Events**: xAPI statements.

---

## 📅 Deliverables by Phase

### Phase 0 — Foundations (Week 1)
- Monorepo scaffold (per `.cursorrules`), CI, pre-commit, root Makefile + workspace scripts.
- `packages/openapi` (spec stub) → generate `packages/openapi-client`.
- Empty studio & player apps booting; Go API skeleton (health, project CRUD).

**Exit Criteria**
- `make dev` starts backend + frontend.
- `GET /api/v1/health` green; frontend `/health` shows API ping.

### Phase 1 — Authoring MVP (Weeks 2–4)
- **Studio**
  - Canvas editor with grid/snapping; Palette with basic question blocks.
  - Property inspector (JSON Forms or rjsf `uiSchema`) for block props.
  - Export **Quiz Bundle**: `ui.json` (Adaptive Cards) + `quiz.json` (QTI item set) + assets.
- **Player**
  - Adaptive Cards renderer + binding layer (map inputs/actions → item responses).
  - Response session state; submit → backend.
- **Backend**
  - Projects, Items, Publish endpoints.
  - Attempts: create/patch/finish; scoring hook stub.
  - xAPI event emission scaffold.

**Exit Criteria**
- Create a quiz visually, publish, open share link, answer a few items; see attempt saved.

### Phase 2 — Scoring & Analytics (Weeks 5–6)
- Implement QTI scoring for supported item types; store `ScoreReport`.
- Emit xAPI statements for `initialized`, `answered`, `completed`, `passed/failed`.
- Dashboard endpoints for aggregates.

**Exit Criteria**
- Answering items updates score & xAPI stream; dashboard shows results.

### Phase 3 — Collaboration & Templates (Weeks 7–8)
- Yjs multi-cursor presence, comments, history.
- Theming + template gallery.
- Accessibility audit (WCAG 2.2 AA) on player interactions.

**Exit Criteria**
- Two users co-edit; templates create pre-themed quizzes; axe checks pass.

> **Future**: LTI 1.3 integration; QTI import; adaptive testing; marketplace.

---

## 🚀 Development Process

### Ground Rules (enforced via `.cursorrules`)
- File placement, naming, docs, tests, a11y checks as defined.
- PRs < 400 LOC; tests & docs required; OpenAPI updated when endpoints change.

### Development Flow
1. **Design the contract first** in `packages/openapi/openapi.yaml`.
2. Generate TS client → `packages/openapi-client`.
3. Implement backend handler (Go/Chi), service, store; table-driven tests.
4. Implement studio UI: component + property panel; wire to OpenAPI client.
5. Implement player mapping: Adaptive Card element ↔ QTI response binding.
6. Add unit tests (Go + Vitest/RTL) and minimal e2e happy path.
7. Update `/docs` and root `README.md`.

### Tooling
- **Backend**: Go 1.22, Chi, Validator, Zerolog/Slog, Testcontainers.
- **Frontend**: Next.js, Tailwind + shadcn/ui, RHF + Zod, Vitest + RTL, axe.
- **Collab**: Yjs + provider.
- **Contracts**: OpenAPI + openapi-typescript.
- **Analytics**: xAPI to LRS (configurable endpoint).

---

## 📦 Declarative Artifacts

### `ui.json` — Adaptive Cards (visual layout)
- **Why**: portable, declarative, vendor-neutral UI JSON with broad support.
- **Structure**: Adaptive Card root with `body` (containers, text, media, inputs) + `actions`.
- **Extensions**: use `data`/`id`/`Action.Submit` `data` payloads to carry **binding keys** (e.g., `itemId`, `responseId`).

### `quiz.json` — QTI 3.0 (assessment semantics)
- **Why**: widely adopted standard for items/tests & result reporting.
- **Structure**: `qti-assessment-item` (per question) + `qti-assessment-test` (assembly).
- **Mapping**: each UI control binds to a QTI interaction (choice, order, hotspot). Scoring uses QTI rules.

### `attempt.json` — Runtime response state
- Collected on the player, posted to backend; summarized to `ScoreReport` and xAPI stream.

---

## 🔗 Binding Model: Marrying UI to Semantics

**Principle:** UI is free-form (design first), semantics are rigorous (scoreable).

- Every UI control (Adaptive Cards `Input.*`) has `data: { itemId, responseKey }`.
- The Player transforms `Action.Submit` payload → QTI response processing.
- Media and decorative elements live only in `ui.json`, not in scoring.

> **Bonus**: **Studio Property Panels** use JSON Forms or rjsf `uiSchema` to give rich editors for block attributes; this accelerates authoring, but does **not** affect the public quiz format.

---

## ✅ Acceptance Criteria (MVP)

- Author can create a quiz with at least **6 item types** (choice, multi-select, ordering, hotspot, text-entry, image-select).
- Publish produces a **Quiz Bundle** (ui.json + quiz.json + assets) retrievable by link.
- Player renders from bundle, collects responses, and posts attempts.
- Backend scores **deterministically**, returns `ScoreReport`.
- xAPI statements are emitted to the configured LRS for start/answer/complete.
- A11y checks: keyboard flow, roles/labels, contrast; axe shows no critical issues.

---

## ⚠️ Risks & Mitigations

### UI Spec Limitations
- **Risk**: Some interactions may exceed stock Adaptive Cards.
- **Mitigation**: Use `Action.Execute` and host-specific renderers; fall back to custom React widgets while keeping **the schema as the source of truth**.

### QTI Mapping Complexity
- **Risk**: Complex mapping between UI and QTI standards.
- **Mitigation**: Support a **core subset** first; ship import/export for that subset; expand iteratively.

### Latency/Heavy Assets
- **Risk**: Performance issues with large media files.
- **Mitigation**: CDN, lazy loading, progressive media; Player runs at the edge.

### Collaboration Conflicts
- **Risk**: Concurrent editing conflicts.
- **Mitigation**: CRDTs via Yjs with snapshot/versioning.

---

## 📋 Next Tasks

### Backend
- `POST /api/v1/projects/:id/publish` → accept bundle, store, return public URL.
- `POST /api/v1/attempts` (start), `PATCH /api/v1/attempts/:id` (responses), `POST /api/v1/attempts/:id/finish` (score).
- xAPI emitter (configurable LRS endpoint + auth).

### Studio
- Blocks: Title, Media, Choice, MultiChoice, TextEntry, Ordering, Hotspot.
- Property panels using JSON Forms `uischema`; generate semantic defaults.
- Exporter: compose `ui.json` (Adaptive Cards) + `quiz.json` (QTI) + assets.

### Player
- Adaptive Card renderer with bindings for `Input.Text`, `Input.ChoiceSet`, `Input.Toggle`, and custom Hotspot.
- Submit pipeline → attempt patch → finish → score display.

### Quality
- Axe automated checks for interactive components; Vitest/RTL unit tests.
- Go table-driven tests for scoring & error mapping; testcontainers for Postgres.

---

## 💻 Example Snippets

### Adaptive Cards (ui.json) — Choice Block
```json
{
  "type": "AdaptiveCard",
  "version": "1.5",
  "body": [
    { "type": "TextBlock", "text": "Which are prime numbers?", "wrap": true },
    {
      "type": "Input.ChoiceSet",
      "isMultiSelect": true,
      "choices": [
        { "title": "2", "value": "2" },
        { "title": "3", "value": "3" },
        { "title": "4", "value": "4" }
      ],
      "id": "resp-primes",
      "label": "Select all that apply",
      "data": { "itemId": "item-primes", "responseKey": "choice" }
    }
  ],
  "actions": [{ "type": "Action.Submit", "title": "Next", "data": { "intent": "submit-item" } }]
}
```

### QTI (quiz.json) — Matching Item
```xml
<qti-assessment-item identifier="item-primes" title="Prime numbers" adaptive="false" time-dependent="false" xmlns="http://www.imsglobal.org/xsd/imsqti_v3p0">
  <response-declaration identifier="RESPONSE" cardinality="multiple" base-type="identifier">
    <correct-response>
      <value>2</value>
      <value>3</value>
    </correct-response>
  </response-declaration>
  <item-body>
    <choice-interaction response-identifier="RESPONSE" shuffle="false" max-choices="0">
      <prompt>Which are prime numbers?</prompt>
      <simple-choice identifier="2">2</simple-choice>
      <simple-choice identifier="3">3</simple-choice>
      <simple-choice identifier="4">4</simple-choice>
    </choice-interaction>
  </item-body>
  <response-processing template="http://www.imsglobal.org/question/qti_v3p0/rptemplates/match_correct"/>
</qti-assessment-item>
```

---

## 📚 Definitions & Links

### Adaptive Cards
Platform-agnostic declarative UI JSON with multiple renderers.
- [adaptivecards.io](https://adaptivecards.io)
- [Microsoft Learn](https://learn.microsoft.com/en-us/adaptive-cards/)

### IMS QTI 3.0
Standard for interoperable assessment items/tests/results.
- [imsglobal.org](https://www.imsglobal.org/)
- [1edtech.org](https://www.1edtech.org/)

### xAPI (Experience API)
Standard for learning telemetry; requires an LRS.
- [docs.rusticisoftware.com](https://docs.rusticisoftware.com/)

### Yjs
CRDT framework for real-time, offline-friendly collaboration.
- [yjs.dev](https://yjs.dev/)

---

## 🎯 Definition of Done

**Done = Shippable**

- ✅ Visual authoring produces standards-based artifacts.
- ✅ Player renders those artifacts faithfully and captures responses.
- ✅ Backend scores and cooperates via stable APIs and events.
- ✅ A11y, tests, and docs are present and passing in CI.
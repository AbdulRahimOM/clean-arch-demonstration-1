## ✅ Clean Architecture Demo

A properly structured Go application following Clean Architecture principles.

### Architecture Layers

```
Handlers (HTTP) → Use Cases → Repositories → Database
      ↓               ↓            ↓
  Framework      Business Logic   Data Access
```

### Key Benefits

1. **Testable** - Business logic tested without HTTP or database
2. **Maintainable** - Clear separation, single responsibility
3. **Flexible** - Swap database or framework easily
4. **No vendor lock-in** - Business logic doesn't know about Fiber or MongoDB

### Project Structure
```
internal/
├── domain/          # Pure business models
├── application/     # Use cases & interfaces
├── infrastructure/  # MongoDB, notifications
└── api/             # HTTP handlers
```

#### Notes
* This project is **for Clean Architecture demonstration purposes only**
* **MongoDB sessions/transactions are intentionally omitted** for simplicity
* Not production-ready or fully error-hardened
* Implemented with AI assistance
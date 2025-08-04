# Load Balancer Assignment

## Overview
You are working with a distributed microservices system. Each service has multiple instances (backends) that need to be load balanced efficiently. Your goal is to implement a **client-side** `Balancer` that distributes incoming requests among these instances. This means the load balancing logic resides on the client and chooses which backend to send the request to — not the server.

You are provided with the following interfaces and helper constructors:

```go
type Request interface{}
type Response interface{}

type Backend interface {
    Invoke(ctx context.Context, req Request) (Response, error)
}

var _ Backend = &BackendImpl{}

// addr contains ip:port of a specific backend instance
func NewBackend(addr string) *BackendImpl

type Balancer struct {
    // TODO
}

var _ Backend = &Balancer{}

// addrs contains the addresses of all backend instances
func NewBalancer(addrs []string) *Balancer {
    // TODO
}
```

The `BackendImpl` is thread-safe.

You must implement the `Balancer` type to satisfy the `Backend` interface.

---

## Levels of Complexity
Start with the **Basic Level**, then continue with **Intermediate** and **Advanced** levels.

### 🔹 Basic Level
- The balancer must implement the `Backend` interface.
- Distribute requests using the **Round-Robin** algorithm.
- Ensure thread safety.
- Handle errors returned from backend calls correctly.

### 🔹 Intermediate Level
- Exclude consistently failing backends from rotation(you are free to use any failing method).
- Return backends to rotation after some time.
- Use well-structured methods and data separation.

### 🔹 Advanced Level
- Distribute load based on how busy each backend is.
- A backend's load is the number of requests it is currently handling.
- You may use Go’s `container/heap` package to efficiently choose the least-loaded backend.

---

## Note on Backend Implementation
You are free to implement your own version of `BackendImpl` for testing and validation purposes. There are no restrictions or preferences on how this is done, as long as it satisfies the required interface:

```go
type Backend interface {
    Invoke(ctx context.Context, req Request) (Response, error)
}
```

Use your creativity to simulate behavior, errors, or different load conditions.

---

## Requirements
- You may add internal fields and logic inside `NewBalancer(...)`.
- Do not modify the provided interfaces.
- Your solution must be thread-safe.
- Implement the logic progressively (Basic → Intermediate → Advanced).

---

## Evaluation Criteria
| Level         | Criteria                                                                 |
|---------------|--------------------------------------------------------------------------|
| Basic         | Round-robin logic, thread-safety, error handling                        |
| Intermediate  | Failure tracking and re-inclusion logic                                 |
| Advanced      | Dynamic load-aware dispatch using concurrent-safe counters or heap      |

---

## Submission
- Submit your implementation as a complete Go package.
- Make sure your solution has no race conditions.

## ❓ Is there any problem?

Please raise an issue or contact your instructor if:
- You’re unsure how to implement the eviction logic
- You need sample test cases or hints

---

### Run
```bash
   golangci-lint run ./...
```

Make sure your code is free of linting errors and warnings before submitting.

---

Good luck!

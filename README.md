# doppel

> Your data's doppelgänger — deep copies without side effects. ✨

`doppel` is a production-grade Go library for **explicit, reflection-free deep cloning** of arbitrary data structures.
Built around a **performance-first, explicit-over-magic** philosophy.

🔗 **Full Documentation**: [docs/INDEX.md](./docs/INDEX.md)

---

## 🚀 Quick Start

```bash
go get github.com/seyallius/doppel
```

```go
package main

import "github.com/seyallius/doppel"

type User struct {
    ID   int64
    Name string
    Tags []string
}

func (u *User) Clone() (*User, error) {
    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }
    return &User{ID: u.ID, Name: u.Name, Tags: tags}, nil
}

func main() {
    original := &User{ID: 1, Name: "Alice", Tags: []string{"admin", "dev"}}
    cloned, _ := doppel.Clone(original) // independent deep copy! ✧◝(⁰▿⁰)◜✧
}
```

---

## 📚 Documentation Map

| Section                 | Description                             | Link                                                     |
|-------------------------|-----------------------------------------|----------------------------------------------------------|
| 🗺️ Index               | Full documentation table of contents    | [docs/INDEX.md](./docs/INDEX.md)                         |
| 🎯 Getting Started      | Why doppel, installation, quick example | [docs/getting-started.md](./docs/getting-started.md)     |
| 💭 Philosophy           | Design principles & priority chain      | [docs/philosophy.md](./docs/philosophy.md)               |
| 🧠 Core Concepts        | `Cloner[T]`, `SelfClonable[T]`, helpers | [docs/core-concepts.md](./docs/core-concepts.md)         |
| 🔧 API Reference        | Complete public API documentation       | [docs/api-reference.md](./docs/api-reference.md)         |
| 🛠️ Usage Guide         | Step-by-step cloning patterns (1-8)     | [docs/usage-guide.md](./docs/usage-guide.md)             |
| ⚙️ Advanced             | Error handling, nil safety, struct tags | [docs/advanced.md](./docs/advanced.md)                   |
| 🔍 Reflection Engine    | Fallback engine & cycle policies        | [docs/reflection-engine.md](./docs/reflection-engine.md) |
| 🧪 Testing & Benchmarks | Test commands, performance results      | [docs/testing.md](./docs/testing.md)                     |
| 🗓️ Roadmap             | Phase breakdown & future plans          | [docs/roadmap.md](./docs/roadmap.md)                     |

---

## 🤝 Contributing

1. Read [docs/INDEX.md](./docs/INDEX.md) for architecture overview
2. Follow the [Design Philosophy](./docs/philosophy.md)
3. Run tests: `go test -race ./...`
4. Open a PR with clear commit messages ✨

---

## 📜 License

MIT — see [LICENSE](./LICENSE) for details.

Made with ❤️ and explicit clone logic. (◕‿◕)

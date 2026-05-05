# doppel

> Your data's doppelgänger — deep copies without side effects. ✨

`doppel` is a production-grade Go library for **explicit, reflection-free deep cloning** of arbitrary data structures.
Built around a **performance-first, explicit-over-magic** philosophy.

🔗 **Full Documentation**: [docs/INDEX.md](./INDEX.md)

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

## 🤝 Contributing

1. Read [docs/INDEX.md](./INDEX.md) for architecture overview
2. Follow the [Design Philosophy](./philosophy.md)
3. Run tests: `go test -race ./...`
4. Open a PR with clear commit messages ✨

---

## 📜 License

MIT — see [LICENSE](../LICENSE) for details.

Made with ❤️ and explicit clone logic. (◕‿◕)

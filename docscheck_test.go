package ordo_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// goBlock matches a fenced Go block in the README.
var goBlock = regexp.MustCompile("(?s)```go\n(.*?)```")

// fragmentHarness supplies the declarations a README fragment is written
// against. The Quickstart block declares Config, UserRepository,
// PostgresUserRepository and UserService; the remaining two types are the
// second implementation and the slice consumer the prose refers to.
const fragmentHarness = `package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/akim0v/ordo"
)

type Config struct{ DSN string }

type UserRepository interface{ Name() string }

type PostgresUserRepository struct{ dsn string }

func NewPostgresUserRepository(cfg *Config) *PostgresUserRepository {
	return &PostgresUserRepository{dsn: cfg.DSN}
}
func (r *PostgresUserRepository) Name() string { return "postgres(" + r.dsn + ")" }

type CacheUserRepository struct{}

func NewCacheUserRepository() *CacheUserRepository { return &CacheUserRepository{} }
func (*CacheUserRepository) Name() string          { return "cache" }

type UserService struct{ repository UserRepository }

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}
func (s *UserService) RepositoryName() string { return s.repository.Name() }

type Report struct{ repositories []UserRepository }

func NewReport(repositories []UserRepository) *Report { return &Report{repositories} }
func (r *Report) Repositories() []UserRepository      { return r.repositories }

// container builds a container a fragment can resolve from.
func container() *ordo.Container {
	c, err := ordo.New(
		ordo.WithValue(&Config{DSN: "localhost"}),
		ordo.WithService[UserRepository](NewPostgresUserRepository),
		ordo.WithKeyedService[UserRepository]("cache", NewCacheUserRepository),
		ordo.WithFactory(NewUserService),
	)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

var _ = errors.Is
var _ = fmt.Println

func main() {
%s
}
`

// TestREADMEBlocksCompile compiles every Go block in the README. A block that
// declares its own package is built as written; anything else is spliced into
// fragmentHarness, which declares the types the surrounding prose introduces.
func TestREADMEBlocksCompile(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README: %v", err)
	}

	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	matches := goBlock.FindAllStringSubmatch(string(readme), -1)
	if len(matches) == 0 {
		t.Fatal("no Go blocks found in README; the extractor is broken")
	}

	for i, m := range matches {
		block := m[1]

		t.Run(blockName(i, block), func(t *testing.T) {
			source := block
			if !strings.HasPrefix(strings.TrimLeft(block, "\n"), "package ") {
				body := block

				// A fragment that resolves from a container without building
				// one is written against the container the prose already
				// established; bind it so the fragment stands alone.
				if strings.Contains(body, "c.") && !strings.Contains(body, "c, err := ordo.New") {
					body = "c := container()\n" + body
				}

				// A fragment that inspects err without producing one is written
				// against a failed New; bind a real registration failure so the
				// classification it shows is the classification it would do.
				if strings.Contains(body, "err") && !strings.Contains(body, "err :=") {
					body = `_, err := ordo.New(ordo.WithFactory("not a function"))` + "\n" + body
				}

				source = strings.Replace(fragmentHarness, "%s", indent(body), 1)
			}

			dir := t.TempDir()
			write(t, filepath.Join(dir, "main.go"), source)
			write(t, filepath.Join(dir, "go.mod"), "module readmeblock\n\ngo 1.27.0\n\n"+
				"require github.com/akim0v/ordo v0.0.0\n\n"+
				"replace github.com/akim0v/ordo => "+root+"\n")

			cmd := exec.Command("go", "build", "./...")
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("README block %d does not compile: %v\n%s\n--- source ---\n%s",
					i, err, out, source)
			}
		})
	}
}

// blockName labels a subtest with the block index and its first line.
func blockName(i int, block string) string {
	first := "empty"
	for _, line := range strings.Split(strings.TrimSpace(block), "\n") {
		if strings.TrimSpace(line) != "" {
			first = strings.TrimSpace(line)
			break
		}
	}

	first = strings.Map(func(r rune) rune {
		if r == ' ' || r == '/' || r == '\t' {
			return '_'
		}
		return r
	}, first)

	if len(first) > 40 {
		first = first[:40]
	}

	return strings.Join([]string{"block", string(rune('0' + i)), first}, "_")
}

// indent shifts a fragment into a function body.
func indent(block string) string {
	lines := strings.Split(strings.TrimRight(block, "\n"), "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) != "" {
			lines[i] = "\t" + l
		}
	}
	return strings.Join(lines, "\n")
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

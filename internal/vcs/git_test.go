package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhouyg1992/agybar/internal/schema"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "init")
	return dir
}

func TestResolveCleanAndDirtyRepo(t *testing.T) {
	dir := initRepo(t)

	got := Resolve(&schema.VCSInfo{Type: "git"}, dir)
	if got == nil {
		t.Fatal("expected non-nil VCSInfo")
	}
	if got.Branch != "main" {
		t.Errorf("Branch = %q, want main", got.Branch)
	}
	if got.Dirty {
		t.Errorf("expected clean repo, got Dirty=true")
	}

	// Subdirectory inside repo should still resolve via parent walk.
	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	// Untracked file makes repo dirty.
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gotSub := Resolve(nil, sub)
	if gotSub == nil || gotSub.Branch != "main" || !gotSub.Dirty {
		t.Errorf("expected main dirty from subdir, got %+v", gotSub)
	}
}

func TestResolveDetachedHead(t *testing.T) {
	dir := initRepo(t)
	sha := runGit(t, dir, "rev-parse", "HEAD")
	runGit(t, dir, "checkout", "--detach", sha)

	got := Resolve(nil, dir)
	want := "@" + sha[:7]
	if got == nil || got.Branch != want {
		t.Errorf("detached Branch = %+v, want %q", got, want)
	}
}

func TestResolveWorktree(t *testing.T) {
	dir := initRepo(t)
	wtDir := filepath.Join(t.TempDir(), "wt")
	runGit(t, dir, "worktree", "add", "-b", "feat/wt", wtDir)

	got := Resolve(nil, wtDir)
	if got == nil || got.Branch != "feat/wt" {
		t.Errorf("worktree Branch = %+v, want feat/wt", got)
	}
}

func TestResolvePreservesExistingAndSkipsNonGit(t *testing.T) {
	dir := initRepo(t)

	pre := &schema.VCSInfo{Type: "git", Branch: "from-agy", Dirty: true}
	if got := Resolve(pre, dir); got != pre {
		t.Errorf("expected pre-populated VCSInfo to be preserved, got %+v", got)
	}

	hg := &schema.VCSInfo{Type: "hg"}
	if got := Resolve(hg, dir); got != hg {
		t.Errorf("expected non-git VCSInfo to be skipped, got %+v", got)
	}

	nonRepo := t.TempDir()
	if got := Resolve(nil, nonRepo); got != nil {
		t.Errorf("expected nil for non-git dir, got %+v", got)
	}
}

package vcs

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhouyg1992/agybar/internal/schema"
)

const dirtyCheckTimeout = 150 * time.Millisecond

// Resolve fills in Git branch and dirty state from dir when existing VCS info
// lacks a branch. Non-git VCS types and already-populated branches are preserved.
func Resolve(existing *schema.VCSInfo, dir string) *schema.VCSInfo {
	if existing != nil && existing.Branch != "" {
		return existing
	}
	if existing != nil && existing.Type != "" && !strings.EqualFold(existing.Type, "git") {
		return existing
	}
	if strings.TrimSpace(dir) == "" {
		return existing
	}

	branch := detectBranch(dir)
	if branch == "" {
		return existing
	}

	dirty := detectDirty(dir)
	return &schema.VCSInfo{
		Type:   "git",
		Branch: branch,
		Dirty:  dirty,
	}
}

func detectBranch(dir string) string {
	gitDir := findGitDir(dir)
	if gitDir == "" {
		return ""
	}

	headBytes, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return ""
	}
	head := strings.TrimSpace(string(headBytes))
	if head == "" {
		return ""
	}

	if strings.HasPrefix(head, "ref:") {
		ref := strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
		if strings.HasPrefix(ref, "refs/heads/") {
			return strings.TrimPrefix(ref, "refs/heads/")
		}
		return ref
	}

	if len(head) >= 7 {
		return "@" + head[:7]
	}
	return head
}

func findGitDir(startDir string) string {
	cur := filepath.Clean(startDir)
	for {
		dotGit := filepath.Join(cur, ".git")
		info, err := os.Stat(dotGit)
		if err == nil {
			if info.IsDir() {
				return dotGit
			}
			// Worktree or submodule: .git is a file containing "gitdir: <path>".
			data, readErr := os.ReadFile(dotGit)
			if readErr == nil {
				line := strings.TrimSpace(string(data))
				if strings.HasPrefix(line, "gitdir:") {
					target := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
					if target != "" {
						if !filepath.IsAbs(target) {
							target = filepath.Join(cur, target)
						}
						return filepath.Clean(target)
					}
				}
			}
			return ""
		}

		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
}

func detectDirty(dir string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), dirtyCheckTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"git",
		"--no-optional-locks",
		"-C", dir,
		"status",
		"--porcelain",
		"--ignore-submodules=dirty",
		"--untracked-files=normal",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	return len(strings.TrimSpace(out.String())) > 0
}

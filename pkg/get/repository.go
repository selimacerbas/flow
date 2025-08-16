package get

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/selimacerbas/flow/internal/common"
)

func GetCommitSHA(repoRoot, ref string) (string, error) {
	var tildeRe = regexp.MustCompile(`^(.*)~(\d+)$`)
	// Handle "...~N" safely.
	if m := tildeRe.FindStringSubmatch(ref); m != nil {
		base := strings.TrimSpace(m[1])
		if base == "" {
			base = "HEAD"
		}
		n, _ := strconv.Atoi(m[2])

		// Resolve base to a commit (handles annotated tags).
		baseOut, err := exec.Command("git", "-C", repoRoot, "rev-parse", "--verify", base+"^{commit}").Output()
		if err != nil {
			return "", fmt.Errorf("resolve base %q: %w", base, err)
		}
		cur := strings.TrimSpace(string(baseOut))
		if cur == "" {
			return "", fmt.Errorf("empty SHA for base %q", base)
		}

		// Walk N first-parent steps; if we run out of parents → ZeroCommit.
		for i := 0; i < n; i++ {
			parOut, err := exec.Command("git", "-C", repoRoot, "show", "-s", "--format=%P", cur).Output()
			if err != nil {
				// Don’t pretend this is root; surface the error (e.g., shallow history).
				return "", fmt.Errorf("read parents of %s: %w", cur, err)
			}
			parents := strings.Fields(strings.TrimSpace(string(parOut)))
			if len(parents) == 0 {
				// We would step past the root commit.
				return common.ZeroCommit, nil
			}
			cur = parents[0] // first parent
		}

		// There are at least N ancestors → resolve the original ref normally.
		out, err := exec.Command("git", "-C", repoRoot, "rev-parse", "--verify", ref+"^{commit}").Output()
		if err != nil {
			return "", fmt.Errorf("rev-parse %q: %w", ref, err)
		}
		sha := strings.TrimSpace(string(out))
		if sha == "" {
			return "", errors.New("empty SHA for " + ref)
		}
		return sha, nil
	}

	// Normal path: resolve any ref/expr to a commit.
	out, err := exec.Command("git", "-C", repoRoot, "rev-parse", "--verify", ref+"^{commit}").Output()
	if err != nil {
		return "", fmt.Errorf("rev-parse %q: %w", ref, err)
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", errors.New("empty SHA for " + ref)
	}
	return sha, nil
}

// MergeBase returns `git merge-base ref branch`.
func GetMergeBase(repoRoot, ref, branch string) (string, error) {
	ref = strings.TrimSpace(ref)
	out, err := exec.Command("git", "-C", repoRoot, "merge-base", ref, branch).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Shorten returns the first n chars of a SHA (or the input if shorter).
func Shorten(sha string, n int) string {
	if n <= 0 {
		n = 7
	}
	if len(sha) <= n {
		return sha
	}
	return sha[:n]
}

func GetChangedDirs(projectRoot, relPath, ref1, ref2 string) ([]string, error) {
	out, err := exec.Command("git", "-C", projectRoot, "diff", "--name-only", ref1, ref2).Output()
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(bytes.TrimSpace(out), []byte{'\n'})
	seen := map[string]struct{}{} // data type set in Python

	for _, line := range lines {
		candidate := string(line)
		if !strings.HasPrefix(candidate, relPath+"/") {
			continue
		}
		candidate = strings.TrimPrefix(candidate, relPath+"/")

		parts := strings.SplitN(candidate, "/", 2)
		top := parts[0]
		seen[top] = struct{}{}
	}

	var dirs []string
	for d := range seen {
		dirs = append(dirs, d)
	}
	return dirs, nil
}

// ensure empty slices encode as [] not null
func NotNull(xs []string) []string {
	if xs == nil {
		return []string{}
	}
	return xs
}

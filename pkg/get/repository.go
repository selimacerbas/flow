package get

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const ZeroCommit = "0000000000000000000000000000000000000000"
const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904" // git's empty tree

var tildeRe = regexp.MustCompile(`^(.*)~(\d+)$`)

// GetCommitSHA resolves a ref (branch/tag/SHA/rev expr) to a commit SHA.
// Special case: if the ref ends with "~N" and the base commit does not have
// N first-parents (i.e., we'd walk past the root), this returns ZeroCommit.
// It does NOT mask shallow-history errors: if a parent can't be read due to
// a shallow clone (or other errors), it returns an error rather than ZeroCommit.
func GetCommitSHA(repoRoot, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		ref = "HEAD"
	}

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
				return ZeroCommit, nil
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
func MergeBase(repoRoot, ref, branch string) (string, error) {
	ref = strings.TrimSpace(ref)
	out, err := exec.Command("git", "-C", repoRoot, "merge-base", ref, branch).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RangeString forms BEFORE..AFTER or BEFORE...AFTER.
func RangeString(before, after string, threeDot bool) string {
	if threeDot {
		return before + "..." + after
	}
	return before + ".." + after
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

// ChangedTopLevelDirs returns a unique, sorted list of top-level directories
// under baseRel (relative to repo root) that changed between before and after.
func ChangedTopLevelDirs(repoRoot, baseRel, before, after string) ([]string, error) {
	base := filepath.ToSlash(filepath.Clean(baseRel))
	if base == "." || base == "/" {
		base = ""
	}

	// Handle "first commit" by diffing the empty tree
	if strings.TrimSpace(before) == "" {
		var err error
		before, err = GetParentOrZero(repoRoot, after)
		if err != nil {
			return nil, err
		}
	}
	if before == ZeroCommit {
		before = emptyTree
	}

	// git diff --name-only BEFORE AFTER
	out, err := exec.Command("git", "-C", repoRoot, "diff", "--name-only", before, after).Output()
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(bytes.TrimSpace(out), []byte{'\n'})
	seen := map[string]struct{}{}

	for _, b := range lines {
		p := string(b)
		if p == "" {
			continue
		}
		p = filepath.ToSlash(filepath.Clean(p))

		// filter by base path
		if base != "" {
			if p == base {
				continue
			}
			if !strings.HasPrefix(p, base+"/") {
				continue
			}
			p = strings.TrimPrefix(p, base+"/")
		}

		// extract first segment (top-level folder under base)
		parts := strings.SplitN(p, "/", 2)
		top := parts[0]
		if top == "" || top == "." {
			continue
		}
		seen[top] = struct{}{}
	}

	var dirs []string
	for d := range seen {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs, nil
}

// JoinRel joins parts into a cleaned relative path.
func JoinRel(parts ...string) string {
	return filepath.Clean(filepath.Join(parts...))
}

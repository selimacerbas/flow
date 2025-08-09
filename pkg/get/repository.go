package get

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const ZeroCommit = "0000000000000000000000000000000000000000"
const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904" // git's empty tree

// GetCommitSHA resolves any ref (branch/tag/SHA/expr like HEAD~2) to a commit SHA.
func GetCommitSHA(repoRoot, ref string) (string, error) {
	ref = strings.TrimSpace(ref)

	out, err := exec.Command("git", "-C", repoRoot, "rev-parse", ref).Output()
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", errors.New("got empty SHA for " + ref)
	}
	return sha, nil
}

// GetParentOrZero returns <ref>~1 or ZeroCommit if there is no parent (first commit).
func GetParentOrZero(repoRoot, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		ref = "HEAD"
	}
	out, err := exec.Command("git", "-C", repoRoot, "rev-parse", ref+"~1").CombinedOutput()
	if err != nil {
		// no parent commit → first push
		return ZeroCommit, nil
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", errors.New("got empty SHA for " + ref + "~1")
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

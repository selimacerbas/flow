package workflowrunsha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/internal/common"
	"github.com/selimacerbas/flow/internal/utils"
)

type LatestSuccessfulWorkflowRunSHAOptions struct {
	GitOwner    string
	GitRepo     string
	GitWorkflow string
	GitBranch   string
	GitToken    string
	Output      string
}

var defaults = &LatestSuccessfulWorkflowRunSHAOptions{
	Output: "text",
}

func init() {
	d := defaults
	f := WorkflowRunSHA.Flags()

    f.StringVar(&d.GitOwner, "git-owner", d.GitOwner, "GitHub owner/org. Reads from config/env.")
    f.StringVar(&d.GitRepo, "git-repo", d.GitRepo, "GitHub repository name. Reads from config/env.")
    f.StringVar(&d.GitWorkflow, "git-workflow", d.GitWorkflow, "Workflow file name under .github/workflows (e.g., build.yaml).")
    f.StringVar(&d.GitBranch, "git-branch", d.GitBranch, "Branch name to query. Reads from env (e.g., GITHUB_REF_NAME).")
    f.StringVar(&d.GitToken, "git-token", d.GitToken, "GitHub token. Reads from config/env (GITHUB_TOKEN or GH_TOKEN).")
    f.StringVarP(&d.Output, "output", "o", d.Output, "Output format (text|json). Default: text")
}

var WorkflowRunSHA = &cobra.Command{
	Use:   "workflowrun-sha",
	Short: "Print the head SHA of the last successful GitHub workflow run on a branch.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		d := defaults

		owner := common.ResolveGitOwner(d.GitOwner)
		repo := common.ResolveGitRepo(d.GitRepo)
		token := common.ResolveGitToken(d.GitToken)
		workflow := common.ResolveGitWorkflow(d.GitWorkflow)
		branch := common.ResolveGitBranch(d.GitBranch)

		if owner == "" || repo == "" || workflow == "" || branch == "" {
			log.Fatalf("--git-owner, --git-repo, --git-workflow and --git-branch are required (pass flags, config, or env)")
		}

		var run *runInfo
		var err error
		switch {
		case token != "":
			// Prefer HTTP when we have a token (works in CI & private repos)
			run, err = fetchHTTP(owner, repo, workflow, branch, token)
			if err != nil {
				log.Fatalf("failed to fetch run info via HTTP %v", err)
			}
		case token == "" && utils.HasBin("gh"):
			// Otherwise, try local gh CLI (uses user's gh auth & host config).
			run, err = fetchViaGH(owner, repo, workflow, branch)
			if err != nil {
				log.Fatalf("failed to fetch run info via GH %v", err)
			}
		default:
			// As a last resort, try unauthenticated HTTP (public repos only).
			run, err = fetchHTTP(owner, repo, workflow, branch, "")
			if err != nil {
				log.Fatalf("failed to fetch run info via HTTP %v", err)
			}
		}

		if run == nil {
			if d.Output == "json" {
				// Return a JSON null to be explicit for machine consumers.
				os.Stdout.WriteString("null\n")
			}
			// text mode: print nothing (CI-friendly) and exit 0
			return
		}

		switch d.Output {
		case "json":
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
				"run_id":      run.ID,
				"head_sha":    run.HeadSHA,
				"run_number":  run.RunNumber,
				"run_attempt": run.RunAttempt,
				"status":      run.Status,
				"conclusion":  run.Conclusion,
				"event":       run.Event,
				"updated_at":  run.UpdatedAt,
				"url":         run.HTMLURL,
			})
		default:
			fmt.Println(run.HeadSHA)
		}
	},
}

// ---- Fetchers ----
type runInfo struct {
	ID         int64     `json:"id"`
	HeadSHA    string    `json:"head_sha"`
	RunNumber  int       `json:"run_number"`
	RunAttempt int       `json:"run_attempt"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	Event      string    `json:"event"`
	UpdatedAt  time.Time `json:"updated_at"`
	HTMLURL    string    `json:"html_url"`
}

type ghListResp struct {
	Runs []runInfo `json:"workflow_runs"`
}

func fetchHTTP(owner, repo, workflow, branch, token string) (*runInfo, error) {
	u := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/actions/workflows/%s/runs?branch=%s&status=success&exclude_pull_requests=true&per_page=1",
		url.PathEscape(owner),
		url.PathEscape(repo),
		url.PathEscape(workflow),
		url.QueryEscape(branch),
	)

	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Note: private repos without token intentionally return 404
		return nil, fmt.Errorf("github api: %s (check owner/repo/workflow name/token)", resp.Status)
	}

	var out ghListResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode github response: %w", err)
	}
	if len(out.Runs) == 0 {
		return nil, nil
	}
	return &out.Runs[0], nil
}

func fetchViaGH(owner, repo, workflow, branch string) (*runInfo, error) {
	args := []string{
		"api",
		fmt.Sprintf("repos/%s/%s/actions/workflows/%s/runs", owner, repo, workflow),
		"--method", "GET", // keep it GET; otherwise -f/-F would POST and 404
		"-H", "X-GitHub-Api-Version: 2022-11-28",
		"-F", "branch=" + branch,
		"-F", "status=success",
		"-F", "exclude_pull_requests=true",
		"-F", "per_page=1",
	}
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh api failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}

	var out ghListResp
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("decode gh api response: %w", err)
	}
	if len(out.Runs) == 0 {
		return nil, nil
	}
	return &out.Runs[0], nil
}

// func fetchViaGH(owner, repo, workflow, branch string) (*runInfo, error) {
// 	// Use the repo-wide endpoint + query string so it's a GET (matches your working shell cmd)
// 	endpoint := fmt.Sprintf(
// 		"/repos/%s/%s/actions/runs?branch=%s&status=success&exclude_pull_requests=true&per_page=1",
// 		owner, repo, url.QueryEscape(branch),
// 	)
//
// 	args := []string{
// 		"api",
// 		endpoint,
// 		"--method", "GET", // harmless here, keeps it explicit
// 		"-H", "X-GitHub-Api-Version: 2022-11-28",
// 	}
//
// 	cmd := exec.Command("gh", args...)
// 	var stdout, stderr bytes.Buffer
// 	cmd.Stdout = &stdout
// 	cmd.Stderr = &stderr
// 	if err := cmd.Run(); err != nil {
// 		return nil, fmt.Errorf("gh api failed: %v: %s", err, strings.TrimSpace(stderr.String()))
// 	}
//
// 	var out ghListResp
// 	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
// 		return nil, fmt.Errorf("decode gh api response: %w", err)
// 	}
// 	if len(out.Runs) == 0 {
// 		return nil, nil
// 	}
// 	return &out.Runs[0], nil
// }

// func fetchViaGH(owner, repo, workflow, branch string) (*runInfo, error) {
// 	args := []string{
// 		"api",
// 		fmt.Sprintf("repos/%s/%s/actions/workflows/%s/runs", owner, repo, workflow),
// 		"--method", "GET", // <- keep it GET; otherwise -f/-F makes it a POST and 404s
// 		"-H", "X-GitHub-Api-Version: 2022-11-28",
// 		"-f", "branch=" + branch,
// 		"-f", "status=success",
// 		"-f", "per_page=1",
// 	}
// 	cmd := exec.Command("gh", args...)
// 	var stdout, stderr bytes.Buffer
// 	cmd.Stdout = &stdout
// 	cmd.Stderr = &stderr
// 	if err := cmd.Run(); err != nil {
// 		return nil, fmt.Errorf("gh api failed: %v: %s", err, strings.TrimSpace(stderr.String()))
// 	}
//
// 	var out ghListResp
// 	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
// 		return nil, fmt.Errorf("decode gh api response: %w", err)
// 	}
// 	if len(out.Runs) == 0 {
// 		return nil, nil
// 	}
// 	return &out.Runs[0], nil
// }

// func resolveDefaults(o *Options) {
// 	// Owner/Repo from Actions env if not provided
// 	if (o.Owner == "" || o.Repo == "") && os.Getenv("GITHUB_REPOSITORY") != "" {
// 		parts := strings.SplitN(os.Getenv("GITHUB_REPOSITORY"), "/", 2)
// 		if len(parts) == 2 {
// 			if o.Owner == "" {
// 				o.Owner = parts[0]
// 			}
// 			if o.Repo == "" {
// 				o.Repo = parts[1]
// 			}
// 		}
// 	}
// 	// Branch from env if not provided
// 	if o.Branch == "" {
// 		o.Branch = os.Getenv("GITHUB_REF_NAME")
// 	}
// 	// Token: prefer your config resolver, then env
// 	if o.GitToken == "" {
// 		o.GitToken = common.ResolveGitToken(o.GitToken) // your helper (reads config/env)
// 	}
// 	if o.GitToken == "" {
// 		if t := os.Getenv("GITHUB_TOKEN"); t != "" {
// 			o.GitToken = t
// 		} else if t := os.Getenv("GH_TOKEN"); t != "" {
// 			o.GitToken = t
// 		}
// 	}
// }

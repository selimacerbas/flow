package workflowstatus

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type opts struct {
	Owner    string
	Repo     string
	Workflow string // e.g. "build.yaml"
	Branch   string // defaults from env
	Output   string // text|json
	Token    string // defaults from env
}

var d = &opts{
	Output: "text",
}

var Cmd = &cobra.Command{
	Use:   "workflow-status",
	Short: "Print head SHA of the last successful run of a workflow on a branch (GitHub).",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		// Defaults from Actions env, so it works flagless in CI except --workflow
		if d.Owner == "" || d.Repo == "" {
			if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
				if parts := strings.SplitN(repo, "/", 2); len(parts) == 2 {
					if d.Owner == "" {
						d.Owner = parts[0]
					}
					if d.Repo == "" {
						d.Repo = parts[1]
					}
				}
			}
		}
		if d.Branch == "" {
			d.Branch = os.Getenv("GITHUB_REF_NAME")
		}
		if d.Token == "" {
			d.Token = os.Getenv("GITHUB_TOKEN")
			if d.Token == "" {
				d.Token = os.Getenv("GH_TOKEN")
			}
		}
		if d.Owner == "" || d.Repo == "" || d.Workflow == "" || d.Branch == "" {
			log.Fatalf("owner, repo, workflow, and branch are required (set flags or GitHub Actions env)")
		}

		// Build request
		u := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/runs?branch=%s&status=success&per_page=1",
			url.PathEscape(d.Owner),
			url.PathEscape(d.Repo),
			url.PathEscape(d.Workflow),
			url.QueryEscape(d.Branch),
		)

		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		if d.Token != "" {
			req.Header.Set("Authorization", "Bearer "+d.Token)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("github api request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Fatalf("github api: %s", resp.Status)
		}

		var out struct {
			Runs []struct {
				ID         int64     `json:"id"`
				HeadSHA    string    `json:"head_sha"`
				RunNumber  int       `json:"run_number"`
				RunAttempt int       `json:"run_attempt"`
				Status     string    `json:"status"`
				Conclusion string    `json:"conclusion"`
				Event      string    `json:"event"`
				UpdatedAt  time.Time `json:"updated_at"`
				HTMLURL    string    `json:"html_url"`
			} `json:"workflow_runs"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			log.Fatalf("decode github response: %v", err)
		}
		if len(out.Runs) == 0 {
			// Minimal behavior: print nothing so callers can fall back cleanly.
			return
		}

		switch d.Output {
		case "json":
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
				"head_sha":    out.Runs[0].HeadSHA,
				"run_id":      out.Runs[0].ID,
				"run_number":  out.Runs[0].RunNumber,
				"run_attempt": out.Runs[0].RunAttempt,
				"status":      out.Runs[0].Status,
				"conclusion":  out.Runs[0].Conclusion,
				"event":       out.Runs[0].Event,
				"updated_at":  out.Runs[0].UpdatedAt,
				"url":         out.Runs[0].HTMLURL,
			})
		default: // text
			fmt.Println(out.Runs[0].HeadSHA)
		}
	},
}

func init() {
	f := Cmd.Flags()
	f.StringVar(&d.Owner, "owner", d.Owner, "GitHub owner (defaults from GITHUB_REPOSITORY)")
	f.StringVar(&d.Repo, "repo", d.Repo, "GitHub repo (defaults from GITHUB_REPOSITORY)")
	f.StringVar(&d.Workflow, "workflow", d.Workflow, "Workflow file name, e.g. build.yaml (required)")
	f.StringVar(&d.Branch, "branch", d.Branch, "Branch name (defaults from GITHUB_REF_NAME)")
	f.StringVarP(&d.Output, "output", "o", d.Output, "Output format: text|json")
	f.StringVar(&d.Token, "token", d.Token, "GitHub token (defaults from GITHUB_TOKEN/GH_TOKEN)")
}

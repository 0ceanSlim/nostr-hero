package staging

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v67/github"
	"golang.org/x/oauth2"
	"pubkey-quest/cmd/codex/config"
)

// GitHubClient handles GitHub API operations
type GitHubClient struct {
	client *github.Client
	cfg    *config.Config
}

// NewGitHubClient creates a new GitHub client with authentication
func NewGitHubClient(cfg *config.Config) *GitHubClient {
	if cfg.GitHub.Token == "" {
		return nil
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHub.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GitHubClient{
		client: github.NewClient(tc),
		cfg:    cfg,
	}
}

// CreatePR creates a pull request from staged changes
func (gc *GitHubClient) CreatePR(session *Session) (prURL string, err error) {
	ctx := context.Background()

	// 1. Get default branch (main or master)
	repo, _, err := gc.client.Repositories.Get(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName)
	if err != nil {
		return "", fmt.Errorf("failed to get repository: %v", err)
	}
	baseBranch := repo.GetDefaultBranch()
	log.Printf("📂 Using base branch: %s", baseBranch)

	// 2. Get latest commit SHA on base branch
	ref, _, err := gc.client.Git.GetRef(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName, "refs/heads/"+baseBranch)
	if err != nil {
		return "", fmt.Errorf("failed to get base branch ref: %v", err)
	}
	baseSHA := ref.Object.GetSHA()
	log.Printf("📍 Base SHA: %s", baseSHA)

	// 3. Create new branch
	branchName := fmt.Sprintf("codex-submissions/%s", session.ID)
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: github.String(baseSHA),
		},
	}
	_, _, err = gc.client.Git.CreateRef(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName, newRef)
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %v", err)
	}
	log.Printf("🌿 Created branch: %s", branchName)

	// 4. Get base tree
	baseCommit, _, err := gc.client.Git.GetCommit(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName, baseSHA)
	if err != nil {
		return "", fmt.Errorf("failed to get base commit: %v", err)
	}
	baseTreeSHA := baseCommit.Tree.GetSHA()

	// 5. Create tree entries for all changes
	var treeEntries []*github.TreeEntry
	for _, change := range session.Changes {
		if change.Type == ChangeDelete {
			// For deletions, create a tree entry with nil SHA
			treeEntries = append(treeEntries, &github.TreeEntry{
				Path: github.String(change.FilePath),
				Mode: github.String("100644"),
				Type: github.String("blob"),
				SHA:  nil, // nil SHA indicates deletion
			})
		} else {
			// For creates/updates, upload the blob
			blob, _, err := gc.client.Git.CreateBlob(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName, &github.Blob{
				Content:  github.String(base64.StdEncoding.EncodeToString(change.NewContent)),
				Encoding: github.String("base64"),
			})
			if err != nil {
				return "", fmt.Errorf("failed to create blob for %s: %v", change.FilePath, err)
			}

			treeEntries = append(treeEntries, &github.TreeEntry{
				Path: github.String(change.FilePath),
				Mode: github.String("100644"),
				Type: github.String("blob"),
				SHA:  blob.SHA,
			})
		}
	}

	// 6. Create tree
	tree, _, err := gc.client.Git.CreateTree(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName, baseTreeSHA, treeEntries)
	if err != nil {
		return "", fmt.Errorf("failed to create tree: %v", err)
	}
	log.Printf("🌳 Created tree: %s", tree.GetSHA())

	// 7. Create commit
	commitMsg := gc.generateCommitMessage(session)
	commit, _, err := gc.client.Git.CreateCommit(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName, &github.Commit{
		Message: github.String(commitMsg),
		Tree:    tree,
		Parents: []*github.Commit{{SHA: github.String(baseSHA)}},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create commit: %v", err)
	}
	log.Printf("💾 Created commit: %s", commit.GetSHA())

	// 8. Update branch reference to point to new commit
	branchRef := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: commit.SHA,
		},
	}
	_, _, err = gc.client.Git.UpdateRef(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName, branchRef, false)
	if err != nil {
		return "", fmt.Errorf("failed to update branch ref: %v", err)
	}

	// 9. Create pull request
	prTitle := gc.generatePRTitle(session)
	prBody := gc.generatePRBody(session)
	pr, _, err := gc.client.PullRequests.Create(ctx, gc.cfg.GitHub.RepoOwner, gc.cfg.GitHub.RepoName, &github.NewPullRequest{
		Title: github.String(prTitle),
		Head:  github.String(branchName),
		Base:  github.String(baseBranch),
		Body:  github.String(prBody),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %v", err)
	}

	log.Printf("✅ Created PR #%d: %s", pr.GetNumber(), pr.GetHTMLURL())
	return pr.GetHTMLURL(), nil
}

// generateCommitMessage creates a descriptive commit message
func (gc *GitHubClient) generateCommitMessage(session *Session) string {
	changeCount := len(session.Changes)
	if changeCount == 1 {
		change := session.Changes[0]
		return fmt.Sprintf("CODEX: %s %s\n\nSubmitted by: %s\nSession: %s",
			strings.Title(string(change.Type)), change.FilePath, session.Npub, session.ID)
	}

	return fmt.Sprintf("CODEX: %d changes to game data\n\nSubmitted by: %s\nSession: %s",
		changeCount, session.Npub, session.ID)
}

// generatePRTitle creates a concise PR title
func (gc *GitHubClient) generatePRTitle(session *Session) string {
	changeCount := len(session.Changes)
	if changeCount == 1 {
		change := session.Changes[0]
		return fmt.Sprintf("CODEX: %s %s", strings.Title(string(change.Type)), change.FilePath)
	}

	return fmt.Sprintf("CODEX: %d changes to game data", changeCount)
}

// generatePRBody creates a detailed PR description
func (gc *GitHubClient) generatePRBody(session *Session) string {
	var body strings.Builder

	body.WriteString("## CODEX Submission\n\n")
	body.WriteString(fmt.Sprintf("**Submitted by:** %s\n", session.Npub))
	body.WriteString(fmt.Sprintf("**Session ID:** `%s`\n\n", session.ID))

	body.WriteString("### Changes\n\n")

	// Group changes by type
	creates := filterChangesByType(session.Changes, ChangeCreate)
	updates := filterChangesByType(session.Changes, ChangeUpdate)
	deletes := filterChangesByType(session.Changes, ChangeDelete)

	if len(creates) > 0 {
		body.WriteString("**Created:**\n")
		for _, change := range creates {
			body.WriteString(fmt.Sprintf("- `%s`\n", change.FilePath))
		}
		body.WriteString("\n")
	}

	if len(updates) > 0 {
		body.WriteString("**Modified:**\n")
		for _, change := range updates {
			body.WriteString(fmt.Sprintf("- `%s`\n", change.FilePath))
		}
		body.WriteString("\n")
	}

	if len(deletes) > 0 {
		body.WriteString("**Deleted:**\n")
		for _, change := range deletes {
			body.WriteString(fmt.Sprintf("- `%s`\n", change.FilePath))
		}
		body.WriteString("\n")
	}

	body.WriteString("---\n")
	body.WriteString("*Generated by CODEX item editor*\n")

	return body.String()
}

// filterChangesByType filters changes by type
func filterChangesByType(changes []Change, changeType ChangeType) []Change {
	var filtered []Change
	for _, change := range changes {
		if change.Type == changeType {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

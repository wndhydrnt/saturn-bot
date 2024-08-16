package git

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
)

type BranchModifiedError struct {
	Checksums []string
}

func (e *BranchModifiedError) Error() string {
	return "branch contains other commits"
}

type GitCommandError struct {
	err      error
	exitCode int
	stderr   string
	stdout   string
}

func (e *GitCommandError) Error() string {
	return fmt.Sprintf("exit code '%d' stderr '%s' stdout '%s'", e.exitCode, e.stderr, e.stdout)
}

func (e *GitCommandError) Unwrap() error {
	return e.err
}

type GitClient interface {
	CommitChanges(msg string) error
	Execute(arg ...string) (string, string, error)
	HasLocalChanges() (bool, error)
	HasRemoteChanges(branchName string) (bool, error)
	Prepare(repo host.Repository, retry bool) (string, error)
	Push(branchName string) error
	UpdateTaskBranch(branchName string, forceRebase bool, repo host.Repository) (bool, error)
}

type Git struct {
	CmdExec func(*exec.Cmd) error // exists to mock calls in unit tests
	EnvVars []string

	cloneOpts        []string
	checkoutDir      string
	dataDir          string
	defaultCommitMsg string
	gitPath          string
	gitUrl           config.ConfigurationGitUrl
	userEmail        string
	userName         string
}

func New(cfg config.Configuration) (*Git, error) {
	envVars, err := createGitEnvVars(cfg)
	if err != nil {
		return nil, fmt.Errorf("create git auth env vars: %w", err)
	}

	return &Git{
		cloneOpts:        cfg.GitCloneOptions,
		dataDir:          *cfg.DataDir,
		defaultCommitMsg: cfg.GitCommitMessage,
		EnvVars:          envVars,
		CmdExec:          execCmd,
		gitPath:          cfg.GitPath,
		gitUrl:           cfg.GitUrl,
		userEmail:        cfg.GitUserEmail(),
		userName:         cfg.GitUserName(),
	}, nil
}

func (g *Git) CommitChanges(msg string) error {
	if msg == "" {
		msg = g.defaultCommitMsg
	}

	_, _, err := g.Execute("add", ".")
	if err != nil {
		return fmt.Errorf("add changes before commit: %w", err)
	}

	_, _, err = g.Execute("commit", "-m", msg)
	if err != nil {
		return fmt.Errorf("commit changes: %w", err)
	}

	return nil
}

func (g *Git) Prepare(repo host.Repository, retry bool) (string, error) {
	checkoutDir := path.Join(g.dataDir, "git", repo.FullName())
	g.checkoutDir = checkoutDir
	logger := log.GitLogger().With("dir", checkoutDir, "repository", repo.FullName())
	_, err := os.Stat(checkoutDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("check if git checkout dir %s exists: %w", checkoutDir, err)
		}

		err = os.MkdirAll(checkoutDir, 0755)
		if err != nil {
			return "", fmt.Errorf("create git checkout dir %s: %w", checkoutDir, err)
		}

		logger.Debug("Cloning repository")
		cloneArgs := append([]string{"clone", g.getCloneUrl(repo), "."}, g.cloneOpts...)
		_, _, err = g.Execute(cloneArgs...)
		if err != nil {
			return "", fmt.Errorf("clone repository %s: %w", repo.FullName(), err)
		}
	} else {
		logger.Debug("Resetting repository")
		err := g.reset(checkoutDir)
		if err != nil {
			return "", err
		}

		_, _, err = g.Execute("checkout", repo.BaseBranch())
		if err != nil {
			if retry {
				log.GitLogger().Error("Base branch does not exist after retry", "baseBranch", repo.BaseBranch())
				return "", fmt.Errorf("checkout base branch: %w", err)
			} else {
				log.GitLogger().Debug("Failure to check out base branch - deleting local repository and triggering another clone", "baseBranch", repo.BaseBranch())
				err := os.RemoveAll(checkoutDir)
				if err != nil {
					return "", fmt.Errorf("failed to remove checkout directory %s: %w", checkoutDir, err)
				}

				return g.Prepare(repo, true)
			}
		}

		log.GitLogger().Debug("Pulling changes into base branch", "repository", repo.FullName())
		_, _, err = g.Execute("pull", "--prune", "origin")
		if err != nil {
			return "", fmt.Errorf("pull changes from remote into base branch: %w", err)
		}
	}

	userName, userEmail, err := g.author(repo)
	if err != nil {
		return "", fmt.Errorf("get git author info: %w", err)
	}

	_, _, err = g.Execute("config", "user.email", userEmail)
	if err != nil {
		return "", fmt.Errorf("set git user email: %w", err)
	}

	_, _, err = g.Execute("config", "user.name", userName)
	if err != nil {
		return "", fmt.Errorf("set git user name: %w", err)
	}

	return checkoutDir, nil
}

func (g *Git) Execute(arg ...string) (string, string, error) {
	cmd := exec.Command(g.gitPath, arg...) // #nosec G204 -- git executable is checked and arguments are controlled by saturn-bot
	cmd.Dir = g.checkoutDir
	cmd.Env = g.EnvVars
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmdLog := strings.Join(arg, " ")
	log.GitLogger().Debug("Executing git", "cmd", cmdLog, "cwd", g.checkoutDir)
	err := g.CmdExec(cmd)
	if err != nil {
		return stdout.String(), stderr.String(), &GitCommandError{err: err, exitCode: cmd.ProcessState.ExitCode(), stderr: stderr.String(), stdout: stdout.String()}
	}

	return stdout.String(), stderr.String(), nil
}

func (g *Git) HasLocalChanges() (bool, error) {
	stdout, _, err := g.Execute("status", "--porcelain=v1")
	if err != nil {
		return false, fmt.Errorf("list local changes in git: %w", err)
	}

	if stdout == "" {
		return false, nil
	}

	return true, nil
}

func (g *Git) HasRemoteChanges(branchName string) (bool, error) {
	branchExists, err := g.branchExistsRemote(branchName)
	if err != nil {
		return false, fmt.Errorf("check if branch exists to check remote changes: %w", err)
	}

	// Remote branch does not exist. Need to push to remote.
	if !branchExists {
		return true, nil
	}

	stdout, _, err := g.Execute("diff", "--name-only", "origin/"+branchName)
	if err != nil {
		return false, fmt.Errorf("diff remote branch: %w", err)
	}

	return strings.TrimSpace(stdout) != "", nil
}

func (g *Git) Push(branchName string) error {
	_, _, err := g.Execute("push", "origin", branchName, "--force", "--set-upstream")
	if err != nil {
		return fmt.Errorf("git push to branch %s failed: %w", branchName, err)
	}

	return nil
}

func (g *Git) UpdateTaskBranch(branchName string, forceRebase bool, repo host.Repository) (bool, error) {
	_, _, err := g.Execute("checkout", repo.BaseBranch())
	if err != nil {
		return false, fmt.Errorf("checkout base branch %s: %w", repo.BaseBranch(), err)
	}

	branchExistsLocal, err := g.branchExistsLocal(branchName)
	if err != nil {
		return false, err
	}

	branchExistsRemote, err := g.branchExistsRemote(branchName)
	if err != nil {
		return false, err
	}

	if !branchExistsLocal {
		log.GitLogger().Debug("Creating branch", "branch", branchName)
		if branchExistsRemote {
			_, _, err := g.Execute("branch", "--track", branchName, "origin/"+branchName)
			if err != nil {
				return false, fmt.Errorf("create git branch with track %s: %w", branchName, err)
			}
		} else {
			_, _, err := g.Execute("branch", branchName)
			if err != nil {
				return false, fmt.Errorf("create git branch %s: %w", branchName, err)
			}
		}
	}

	hasMergeConflict, err := g.hasMergeConflict(branchName)
	if err != nil {
		return false, err
	}

	log.GitLogger().Debug("Checking out work branch", "branch", branchName)
	_, _, err = g.Execute("checkout", branchName)
	if err != nil {
		return false, fmt.Errorf("checkout git branch %s: %w", branchName, err)
	}

	if branchExistsRemote {
		log.GitLogger().Debug("Pulling changes into work branch", "branch", branchName)
		// --rebase to end up with a clean history.
		// "--strategy-option theirs" to always prefer changes from the remote.
		// Commits by someone else will be preserved with this strategy and there
		// will be no conflict.
		_, _, err = g.Execute("pull", "origin", branchName, "--rebase", "--strategy-option", "theirs")
		if err != nil {
			return false, fmt.Errorf("pull remote changes into git branch %s: %w", branchName, err)
		}
	}

	mergeBase, _, err := g.Execute("merge-base", repo.BaseBranch(), branchName)
	if err != nil {
		return false, fmt.Errorf("pull remote changes into git branch %s: %w", branchName, err)
	}

	mergeBase = strings.TrimSpace(mergeBase)
	if !forceRebase {
		commits, err := g.listForeignCommits(mergeBase, repo)
		if err != nil {
			return false, fmt.Errorf("failed to detect foreign commits: %w", err)
		}

		if len(commits) > 0 {
			return false, &BranchModifiedError{Checksums: commits}
		}
	}

	log.GitLogger().Debug("Resetting to merge base", "branch", branchName)
	_, _, err = g.Execute("reset", "--hard", mergeBase)
	if err != nil {
		return false, fmt.Errorf("reset git branch %s to merge base %s: %w", branchName, mergeBase, err)
	}

	log.GitLogger().Debug("Rebasing onto work branch", "branch", branchName)
	_, _, err = g.Execute("rebase", repo.BaseBranch())
	if err != nil {
		return false, fmt.Errorf("rebase git branch %s: %w", branchName, err)
	}

	return hasMergeConflict, nil
}

func (g *Git) author(repo host.Repository) (string, string, error) {
	if g.userEmail != "" && g.userName != "" {
		return g.userName, g.userEmail, nil
	}

	userInfo, err := repo.Host().AuthenticatedUser()
	if err != nil {
		return "", "", fmt.Errorf("get authenticated user: %w", err)
	}

	if userInfo == nil {
		return "", "", fmt.Errorf("no git author set in configuration and host returned no data for current user")
	}

	return userInfo.Name, userInfo.Email, nil
}

func (g *Git) branchExistsLocal(branchName string) (bool, error) {
	stdout, _, err := g.Execute("branch", "--format", "%(refname)")
	if err != nil {
		return false, fmt.Errorf("check that branch %s exists locally: %w", branchName, err)
	}

	for _, line := range strings.Split(stdout, "\n") {
		search := "refs/heads/" + branchName
		if line == search {
			return true, nil
		}
	}

	return false, nil
}

func (g *Git) branchExistsRemote(branchName string) (bool, error) {
	stdout, _, err := g.Execute("branch", "-r", "--format", "%(refname)")
	if err != nil {
		return false, fmt.Errorf("check that branch %s exists in remote: %w", branchName, err)
	}

	for _, line := range strings.Split(stdout, "\n") {
		search := "refs/remotes/origin/" + branchName
		if line == search {
			return true, nil
		}
	}

	return false, nil
}

func (g *Git) listForeignCommits(mergeBase string, repo host.Repository) ([]string, error) {
	stdout, _, err := g.Execute("rev-list", mergeBase+"..HEAD")
	if err != nil {
		return nil, fmt.Errorf("list rev since merge base: %w", err)
	}

	commitHashesRaw := strings.TrimSpace(stdout)
	if commitHashesRaw == "" {
		return nil, nil
	}

	_, userEmail, err := g.author(repo)
	if err != nil {
		return nil, fmt.Errorf("get git author to check for foreign commits: %w", err)
	}

	var foreignCommits []string
	for _, commitHash := range strings.Split(commitHashesRaw, "\n") {
		commitHash = strings.TrimSpace(commitHash)
		stdout, _, err := g.Execute("show", "--format=%aE", "--no-patch", commitHash)
		if err != nil {
			return nil, fmt.Errorf("show author of commit %s: %w", commitHash, err)
		}

		authorEmail := strings.TrimSpace(stdout)
		if authorEmail != userEmail {
			foreignCommits = append(foreignCommits, commitHash)
		}
	}

	return foreignCommits, nil
}

func (g *Git) hasMergeConflict(branchName string) (bool, error) {
	detected := false
	// Try to merge. Errors if there is a merge conflict.
	_, _, err := g.Execute("merge", branchName, "--no-ff", "--no-commit")
	if err != nil {
		gitErr := &GitCommandError{}
		if errors.As(err, &gitErr) {
			// Exit codes "1" or "2" indicate that a merge is not successful and a conflict exists
			if gitErr.exitCode == 1 || gitErr.exitCode == 2 {
				detected = true
			} else {
				return false, fmt.Errorf("check for merge conflict of branch %s: %w", branchName, err)
			}
		} else {
			return false, fmt.Errorf("unexpected error during check for merge conflict of branch %s: %w", branchName, err)
		}
	}

	// Abort the merge to not leave the branch in a conflicted state
	_, _, err = g.Execute("merge", "--abort")
	if err != nil {
		gitErr := &GitCommandError{}
		if errors.As(err, &gitErr) {
			// 128 is the exit code of the git command if no abort was needed
			if gitErr.exitCode != 128 {
				return false, fmt.Errorf("abort check for merge conflict of branch %s: %w", branchName, err)
			}
		} else {
			return false, fmt.Errorf("unexpected error during abort check for merge conflict of branch %s: %w", branchName, err)
		}
	}

	return detected, nil
}

func (g *Git) getCloneUrl(repo host.Repository) string {
	if g.gitUrl == "ssh" {
		return repo.CloneUrlSsh()
	}

	return repo.CloneUrlHttp()
}

func (g *Git) reset(checkoutDir string) error {
	_, _, err := g.Execute("reset", "--hard")
	if err != nil {
		return fmt.Errorf("reset git checkout %s: %w", checkoutDir, err)
	}

	_, _, err = g.Execute("clean", "-d", "--force")
	if err != nil {
		return fmt.Errorf("clean git checkout %s: %w", checkoutDir, err)
	}

	return nil
}

// Set up authentication for git via environment variables
// See https://git-scm.com/docs/git-config#Documentation/git-config.txt-GITCONFIGCOUNT
func createGitEnvVars(c config.Configuration) ([]string, error) {
	count := 0
	envVars := []string{}
	if c.GithubToken != nil {
		var addr string
		if c.GithubAddress == nil {
			addr = "https://github.com/"
		} else {
			addr = *c.GithubAddress
		}

		u, err := url.Parse(addr)
		if err != nil {
			return nil, fmt.Errorf("parse URL of GitHub: %w", err)
		}

		envVars = append(envVars, []string{
			fmt.Sprintf("GIT_CONFIG_KEY_%d=url.%s://%s@%s/.insteadOf", count, u.Scheme, *c.GithubToken, u.Host),
			fmt.Sprintf("GIT_CONFIG_VALUE_%d=%s", count, addr),
		}...)
		count += 1
	}

	if c.GitlabToken != nil {
		addr := c.GitlabAddress
		if addr == "" {
			addr = "https://gitlab.com/"
		}

		u, err := url.Parse(addr)
		if err != nil {
			return nil, fmt.Errorf("parse URL of GitLab: %w", err)
		}

		envVars = append(envVars, []string{
			fmt.Sprintf("GIT_CONFIG_KEY_%d=url.%s://gitlab-ci-token:%s@%s/.insteadOf", count, u.Scheme, *c.GitlabToken, u.Host),
			fmt.Sprintf("GIT_CONFIG_VALUE_%d=%s", count, addr),
		}...)
		count += 1
	}

	envVars = append(envVars, fmt.Sprintf("GIT_CONFIG_COUNT=%d", count))
	return envVars, nil
}

func execCmd(cmd *exec.Cmd) error {
	return cmd.Run()
}

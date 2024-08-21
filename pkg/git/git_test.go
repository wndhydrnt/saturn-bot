package git_test

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"go.uber.org/mock/gomock"
)

func init() {
	log.InitLog("console", "debug", "debug")
}

func setupOpts(cfg config.Configuration) options.Opts {
	opts := options.Opts{Config: cfg}
	_ = options.Initialize(&opts)
	return opts
}

type execCall struct {
	cmd    *exec.Cmd
	err    error
	stdout *string
}

func (ec *execCall) withDir(dir string) *execCall {
	ec.cmd.Dir = dir
	return ec
}

func (ec *execCall) withErrorMsg(msg string) *execCall {
	ec.err = errors.New(msg)
	return ec
}

func (ec *execCall) withStdout(text string) *execCall {
	ec.stdout = &text
	return ec
}

type execMock struct {
	calls []*execCall
	t     *testing.T
}

func (m *execMock) withCall(name string, args ...string) *execCall {
	ec := &execCall{
		cmd: exec.Command(name, args...),
	}
	m.calls = append(m.calls, ec)
	return ec
}

func (m *execMock) exec(c *exec.Cmd) error {
	if len(m.calls) == 0 {
		m.t.Fatalf("unknown call [%s | %s]", strings.Join(c.Args, ","), c.Dir)
		return nil
	}

	call := m.calls[0]
	if reflect.DeepEqual(call.cmd.Args, c.Args) && call.cmd.Dir == c.Dir {
		m.calls = append(m.calls[:0], m.calls[1:]...)
		if call.stdout != nil {
			_, _ = c.Stdout.Write([]byte(*call.stdout))
		}

		return call.err
	}

	m.t.Fatalf(
		"expected call [%s | %s], but got call [%s | %s]",
		strings.Join(call.cmd.Args, ","),
		call.cmd.Dir,
		strings.Join(c.Args, ","),
		c.Dir,
	)
	return nil
}

func (m *execMock) finished() bool {
	for _, c := range m.calls {
		m.t.Errorf("missing call [%s | %s]", strings.Join(c.cmd.Args, ","), c.cmd.Dir)
	}

	return len(m.calls) == 0
}

func TestGit_Prepare_CloneRepository(t *testing.T) {
	dataDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(dataDir)
		if err != nil {
			panic(err)
		}
	}()

	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()
	repo.EXPECT().CloneUrlHttp().Return("https://git.local/unit/test.git")
	em := &execMock{t: t}
	dir := dataDir + "/git/git.local/unit/test"
	em.withCall("git", "clone", "https://git.local/unit/test.git", ".").withDir(dir)
	em.withCall("git", "config", "user.email", "unit@test.local").withDir(dir)
	em.withCall("git", "config", "user.name", "unittest").withDir(dir)

	cfg := config.Configuration{
		DataDir:   &dataDir,
		GitPath:   "git",
		GitAuthor: "unittest <unit@test.local>",
	}
	g, err := git.New(setupOpts(cfg))
	require.NoError(t, err)
	g.CmdExec = em.exec
	out, err := g.Prepare(repo, false)

	require.NoError(t, err)
	assert.Equal(t, dir, out)
	assert.True(t, em.finished())
}

func TestGit_Prepare_CloneRepositorySsh(t *testing.T) {
	dataDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(dataDir)
		if err != nil {
			panic(err)
		}
	}()

	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()
	repo.EXPECT().CloneUrlSsh().Return("git@git.local/unit/test.git")
	em := &execMock{t: t}
	dir := dataDir + "/git/git.local/unit/test"
	em.withCall("git", "clone", "git@git.local/unit/test.git", ".").withDir(dir)
	em.withCall("git", "config", "user.email", "unit@test.local").withDir(dir)
	em.withCall("git", "config", "user.name", "unittest").withDir(dir)

	cfg := config.Configuration{
		DataDir:   &dataDir,
		GitPath:   "git",
		GitAuthor: "unittest <unit@test.local>",
		GitUrl:    config.ConfigurationGitUrlSsh,
	}
	g, err := git.New(setupOpts(cfg))
	require.NoError(t, err)
	g.CmdExec = em.exec
	out, err := g.Prepare(repo, false)

	require.NoError(t, err)
	assert.Equal(t, dir, out)
	assert.True(t, em.finished())
}

func TestGit_Prepare_UpdateExistingRepository(t *testing.T) {
	dataDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(dataDir)
		if err != nil {
			panic(err)
		}
	}()
	dir := dataDir + "/git/git.local/unit/test"
	err = os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()
	em := &execMock{t: t}
	em.withCall("git", "reset", "--hard").withDir(dir)
	em.withCall("git", "clean", "-d", "--force").withDir(dir)
	em.withCall("git", "checkout", "main").withDir(dir)
	em.withCall("git", "pull", "--prune", "origin").withDir(dir)
	em.withCall("git", "config", "user.email", "unit@test.local").withDir(dir)
	em.withCall("git", "config", "user.name", "unittest").withDir(dir)

	cfg := config.Configuration{
		DataDir:   &dataDir,
		GitPath:   "git",
		GitAuthor: "unittest <unit@test.local>",
	}
	g, err := git.New(setupOpts(cfg))
	require.NoError(t, err)
	g.CmdExec = em.exec
	out, err := g.Prepare(repo, false)

	require.NoError(t, err)
	assert.Equal(t, dir, out)
	assert.True(t, em.finished())
}

func TestGit_Prepare_RetryOnCheckoutError(t *testing.T) {
	dataDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(dataDir)
		if err != nil {
			panic(err)
		}
	}()
	dir := dataDir + "/git/git.local/unit/test"
	err = os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()
	repo.EXPECT().CloneUrlHttp().Return("https://git.local/unit/test.git")
	em := &execMock{t: t}
	em.withCall("git", "reset", "--hard").withDir(dir)
	em.withCall("git", "clean", "-d", "--force").withDir(dir)
	em.withCall("git", "checkout", "main").withDir(dir).withErrorMsg("checkout failed")
	em.withCall("git", "clone", "https://git.local/unit/test.git", ".").withDir(dir)
	em.withCall("git", "config", "user.email", "unit@test.local").withDir(dir)
	em.withCall("git", "config", "user.name", "unittest").withDir(dir)

	g, err := git.New(setupOpts(config.Configuration{
		DataDir:   &dataDir,
		GitPath:   "git",
		GitAuthor: "unittest <unit@test.local>",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	out, err := g.Prepare(repo, false)

	require.NoError(t, err)
	assert.Equal(t, dir, out)
	assert.True(t, em.finished())
}

func TestGit_Prepare_EmptyGitUser(t *testing.T) {
	dataDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(dataDir)
		if err != nil {
			panic(err)
		}
	}()

	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()
	repo.EXPECT().CloneUrlHttp().Return("https://git.local/unit/test.git")
	em := &execMock{t: t}
	dir := dataDir + "/git/git.local/unit/test"
	em.withCall("git", "clone", "https://git.local/unit/test.git", ".").withDir(dir)

	g, err := git.New(setupOpts(config.Configuration{
		DataDir:   toPtr(dataDir),
		GitPath:   "git",
		GitAuthor: "",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	out, err := g.Prepare(repo, false)

	require.NoError(t, err)
	assert.Equal(t, dir, out)
	assert.True(t, em.finished())
}

func TestGit_HasLocal_Changes_Changes(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "status", "--porcelain=v1").withStdout("M  test.txt\n")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	result, err := g.HasLocalChanges()

	require.NoError(t, err)
	require.True(t, result)
	assert.True(t, em.finished())
}

func TestGit_HasLocal_Changes_NoChanges(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "status", "--porcelain=v1").withStdout("")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	result, err := g.HasLocalChanges()

	require.NoError(t, err)
	require.False(t, result)
	assert.True(t, em.finished())
}

func TestGit_HasLocal_Changes_Error(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "status", "--porcelain=v1").withErrorMsg("status failed")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	result, err := g.HasLocalChanges()

	require.Error(t, err)
	require.False(t, result)
	assert.True(t, em.finished())
}

func TestGit_HasRemoteChanges_Changes(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "branch", "-r", "--format", "%(refname)").withStdout("refs/remotes/origin/unittest")
	em.withCall("git", "diff", "--name-only", "origin/unittest").withStdout("test.txt\n")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	result, err := g.HasRemoteChanges("unittest")

	require.NoError(t, err)
	require.True(t, result)
	assert.True(t, em.finished())
}

func TestGit_HasRemoteChanges_BranchDoesNotExist(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "branch", "-r", "--format", "%(refname)").withStdout("refs/remotes/origin/other\nrefs/remotes/origin/another\n")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	result, err := g.HasRemoteChanges("unittest")

	require.NoError(t, err)
	require.True(t, result)
	assert.True(t, em.finished())
}

func TestGit_HasRemoteChanges_ErrorBranchCheck(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "branch", "-r", "--format", "%(refname)").withErrorMsg("branch check failed")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	result, err := g.HasRemoteChanges("unittest")

	assert.Error(t, err)
	require.False(t, result)
	assert.True(t, em.finished())
}

func TestGit_HasRemoteChanges_ErrorDiff(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "branch", "-r", "--format", "%(refname)").withStdout("refs/remotes/origin/unittest")
	em.withCall("git", "diff", "--name-only", "origin/unittest").withErrorMsg("diff failed")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	result, err := g.HasRemoteChanges("unittest")

	assert.Error(t, err)
	require.False(t, result)
	assert.True(t, em.finished())
}

func TestGit_Push_Success(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "push", "origin", "unittest", "--force", "--set-upstream")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	err = g.Push("unittest")

	assert.NoError(t, err)
	assert.True(t, em.finished())
}

func TestGit_Push_Failure(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "push", "origin", "unittest", "--force", "--set-upstream").withErrorMsg("push failed")

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	err = g.Push("unittest")

	assert.Error(t, err)
	assert.True(t, em.finished())
}

func TestGit_UpdateTaskBranch_NewBranch(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "checkout", "main")
	em.withCall("git", "branch", "--format", "%(refname)")
	em.withCall("git", "branch", "-r", "--format", "%(refname)")
	em.withCall("git", "branch", "unittest")
	em.withCall("git", "merge", "unittest", "--no-ff", "--no-commit")
	em.withCall("git", "merge", "--abort")
	em.withCall("git", "checkout", "unittest")
	em.withCall("git", "merge-base", "main", "unittest").withStdout("abc123")
	em.withCall("git", "rev-list", "abc123..HEAD")
	em.withCall("git", "reset", "--hard", "abc123")
	em.withCall("git", "rebase", "main")
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	conflict, err := g.UpdateTaskBranch("unittest", false, repo)

	require.NoError(t, err)
	assert.False(t, conflict)
	assert.True(t, em.finished())
}

func TestGit_UpdateTaskBranch_BranchExistsRemote(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "checkout", "main")
	em.withCall("git", "branch", "--format", "%(refname)")
	em.withCall("git", "branch", "-r", "--format", "%(refname)").withStdout("refs/remotes/origin/main\nrefs/remotes/origin/unittest\n")
	em.withCall("git", "branch", "--track", "unittest", "origin/unittest")
	em.withCall("git", "merge", "unittest", "--no-ff", "--no-commit")
	em.withCall("git", "merge", "--abort")
	em.withCall("git", "checkout", "unittest")
	em.withCall("git", "pull", "origin", "unittest", "--rebase", "--strategy-option", "theirs")
	em.withCall("git", "merge-base", "main", "unittest").withStdout("abc123")
	em.withCall("git", "rev-list", "abc123..HEAD")
	em.withCall("git", "reset", "--hard", "abc123")
	em.withCall("git", "rebase", "main")
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()

	g, err := git.New(setupOpts(config.Configuration{
		DataDir: toPtr("/tmp"),
		GitPath: "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	conflict, err := g.UpdateTaskBranch("unittest", false, repo)

	require.NoError(t, err)
	assert.False(t, conflict)
	assert.True(t, em.finished())
}

func TestGit_UpdateTaskBranch_BranchModified(t *testing.T) {
	em := &execMock{t: t}
	em.withCall("git", "checkout", "main")
	em.withCall("git", "branch", "--format", "%(refname)")
	em.withCall("git", "branch", "-r", "--format", "%(refname)").withStdout("refs/remotes/origin/main\nrefs/remotes/origin/unittest\n")
	em.withCall("git", "branch", "--track", "unittest", "origin/unittest")
	em.withCall("git", "merge", "unittest", "--no-ff", "--no-commit")
	em.withCall("git", "merge", "--abort")
	em.withCall("git", "checkout", "unittest")
	em.withCall("git", "pull", "origin", "unittest", "--rebase", "--strategy-option", "theirs")
	em.withCall("git", "merge-base", "main", "unittest").withStdout("abc123")
	em.withCall("git", "rev-list", "abc123..HEAD").withStdout("a1b2c3d4\n")
	em.withCall("git", "show", "--format=%aE", "--no-patch", "a1b2c3d4").withStdout("user@test.local\n")
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()

	g, err := git.New(setupOpts(config.Configuration{
		DataDir:   toPtr("/tmp"),
		GitAuthor: "unittest <unit@test.local>",
		GitPath:   "git",
	}))
	require.NoError(t, err)
	g.CmdExec = em.exec
	conflict, err := g.UpdateTaskBranch("unittest", false, repo)

	var expectedErr *git.BranchModifiedError
	assert.ErrorAs(t, err, &expectedErr)
	assert.False(t, conflict)
	assert.True(t, em.finished())
}

func Test_New_EnvVars(t *testing.T) {
	testCases := []struct {
		name string
		in   config.Configuration
		want []string
	}{
		{
			name: "GitHub",
			in: config.Configuration{
				GithubToken: toPtr("gh-123"),
			},
			want: []string{
				"GIT_CONFIG_KEY_0=url.https://gh-123@github.com/.insteadOf",
				"GIT_CONFIG_VALUE_0=https://github.com/",
				"GIT_CONFIG_COUNT=1",
			},
		},
		{
			name: "GitLab",
			in: config.Configuration{
				GitlabToken: toPtr("gl-456"),
			},
			want: []string{
				"GIT_CONFIG_KEY_0=url.https://gitlab-ci-token:gl-456@gitlab.com/.insteadOf",
				"GIT_CONFIG_VALUE_0=https://gitlab.com/",
				"GIT_CONFIG_COUNT=1",
			},
		},
		{
			name: "GitHub and GitLab",
			in: config.Configuration{
				GithubToken: toPtr("gh-123"),
				GitlabToken: toPtr("gl-456"),
			},
			want: []string{
				"GIT_CONFIG_KEY_0=url.https://gh-123@github.com/.insteadOf",
				"GIT_CONFIG_VALUE_0=https://github.com/",
				"GIT_CONFIG_KEY_1=url.https://gitlab-ci-token:gl-456@gitlab.com/.insteadOf",
				"GIT_CONFIG_VALUE_1=https://gitlab.com/",
				"GIT_CONFIG_COUNT=2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.in.DataDir = toPtr("/tmp")
			g, err := git.New(options.Opts{Config: tc.in})
			require.NoError(t, err)

			assert.ElementsMatch(t, tc.want, g.EnvVars)
		})
	}
}

func toPtr[T any](v T) *T {
	return &v
}

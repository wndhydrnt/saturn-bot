package host

type repositoryProxy struct {
	Repository

	cache *PullRequestCache
}

func (rp *repositoryProxy) FindPullRequest(branch string) (any, error) {
	pr := rp.cache.Get(branch, rp.FullName())
	if pr != nil {
		return pr.Raw, nil
	}

	return rp.Repository.FindPullRequest(branch)
}

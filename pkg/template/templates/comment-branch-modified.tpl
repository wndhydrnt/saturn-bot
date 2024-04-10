:warning: **This pull request has been modified.**

This is a safety mechanism to prevent saturn-sync from accidentally overriding custom commits.

saturn-sync will not be able to resolve merge conflicts with `{{ .DefaultBranch }}` automatically.
It will not update this pull request or auto-merge it.

Check the box in the description of this PR to force a rebase. This will remove all commits not made by saturn-sync.

The commit(s) that modified the pull request:
{{ range .Checksums }}
- {{ . }}
{{ end }}

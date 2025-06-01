package host

func toSbPr(gpr any) *PullRequest {
	return &PullRequest{
		Raw: gpr,
	}
}

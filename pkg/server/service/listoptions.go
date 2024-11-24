package service

type ListOptions struct {
	Limit int
	Page  int
}

func (lo ListOptions) Offset() int {
	return (lo.Limit * lo.Page) - lo.Limit
}

func (lo ListOptions) Next(count int) int {
	if (lo.Limit * lo.Page) < count {
		return lo.Page + 1
	}

	return 0
}

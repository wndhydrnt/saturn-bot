package service

import "math"

type ListOptions struct {
	Limit int
	Page  int

	total int
}

func (lo ListOptions) Offset() int {
	return (lo.Limit * lo.Page) - lo.Limit
}

func (lo ListOptions) Previous() int {
	if lo.Page == 1 {
		return 0
	}

	return lo.Page - 1
}

func (lo ListOptions) Next() int {
	if (lo.Limit * lo.Page) < lo.total {
		return lo.Page + 1
	}

	return 0
}

func (lo ListOptions) TotalPages() int {
	return int(math.Ceil(float64(lo.total) / float64(lo.Limit)))
}

func (lo *ListOptions) SetTotalItems(t int) {
	lo.total = t
}

func (lo ListOptions) TotalItems() int {
	return lo.total
}

package utils

// ClampPagination normalizes page and pageSize to valid ranges.
// maxPageSize of 0 means no upper limit.
func ClampPagination(page, pageSize, maxPageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if maxPageSize > 0 && pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// CalcTotalPages computes the number of pages for a given total and pageSize.
// Always returns at least 1.
func CalcTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 1
	}
	tp := int(total) / pageSize
	if int(total)%pageSize > 0 {
		tp++
	}
	if tp == 0 {
		tp = 1
	}
	return tp
}

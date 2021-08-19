package s3_uploader

func needExcludeFolder(name string) bool {
	switch name {
	case
		"tests",
		"venv",
		".idea",
		".pytest_cache",
		"__pycache__",
		"original_repo",
		".git":
		return true

	}
	return false
}

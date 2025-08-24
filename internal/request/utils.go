package request

func validateHttpVersion(version string) bool {
	return version == "1.1"
}

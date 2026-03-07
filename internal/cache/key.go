package cache



func GenerateCacheKey(method, url string) string {
	return method + ":" + url
}
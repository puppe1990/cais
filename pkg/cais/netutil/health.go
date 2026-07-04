package netutil

// HealthPayload builds a JSON-serializable /health response with LAN URLs for mobile testing.
func HealthPayload(status, port string) map[string]any {
	return map[string]any{
		"status":   status,
		"lan_urls": LANURLs(port),
	}
}

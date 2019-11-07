package config

import "os"

type Config struct {
	OpenShiftHost string

	MDCContainerName        string
	OauthProxyContainerName string

	UnifiedPushDocumentationURL        string
	IdentityManagementDocumentationURL string
	DataSyncDocumentationURL           string
	MobileSecurityDocumentationURL     string
}

func New() Config {
	return Config{
		OpenShiftHost: getReqEnv("OPENSHIFT_HOST"),

		MDCContainerName:        getEnv("MDC_CONTAINER_NAME", "mdc"),
		OauthProxyContainerName: getEnv("OAUTH_PROXY_CONTAINER_NAME", "mdc-oauth-proxy"),

		// override the default links displayed in MDC for each of the mobile services
		UnifiedPushDocumentationURL:        getEnv("UPS_DOCUMENTATION_URL", "https://docs.aerogear.org/limited-availability/upstream/ups.html"),
		IdentityManagementDocumentationURL: getEnv("IDM_DOCUMENTATION_URL", "https://docs.aerogear.org/limited-availability/upstream/idm.html"),
		DataSyncDocumentationURL:           getEnv("SYNC_DOCUMENTATION_URL", "https://docs.aerogear.org/limited-availability/upstream/sync.html"),
		MobileSecurityDocumentationURL:     getEnv("MSS_DOCUMENTATION_URL", "https://docs.aerogear.org/limited-availability/upstream/mss.html"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getReqEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	panic("Required env var is missing: " + key)
}

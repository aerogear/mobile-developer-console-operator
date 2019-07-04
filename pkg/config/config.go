package config

import (
	"fmt"
	"github.com/aerogear/mobile-developer-console-operator/pkg/util"
	"os"
)

type Config struct {
	OpenShiftHost string

	MDCContainerName        string
	OauthProxyContainerName string

	MDCImageStreamName        string
	MDCImageStreamTag         string
	OauthProxyImageStreamName string
	OauthProxyImageStreamTag  string

	MDCImageStreamInitialImage        string
	OauthProxyImageStreamInitialImage string

	OAuthClientIdSuffix string
	OAuthClientSecret   string
}

func New() Config {
	defaultOAuthClientSecret, err := util.GeneratePassword()
	if err != nil {
		panic(fmt.Sprintf("error generating cookie secret. error : %v", err))
	}

	return Config{
		OpenShiftHost: getReqEnv("OPENSHIFT_HOST"),

		MDCContainerName:        getEnv("MDC_CONTAINER_NAME", "mdc"),
		OauthProxyContainerName: getEnv("OAUTH_PROXY_CONTAINER_NAME", "mdc-oauth-proxy"),

		MDCImageStreamName:        getEnv("MDC_IMAGE_STREAM_NAME", "mdc-imagestream"),
		MDCImageStreamTag:         getEnv("MDC_IMAGE_STREAM_TAG", "latest"),
		OauthProxyImageStreamName: getEnv("OAUTH_PROXY_IMAGE_STREAM_NAME", "mdc-oauth-proxy-imagestream"),
		OauthProxyImageStreamTag:  getEnv("OAUTH_PROXY_IMAGE_STREAM_TAG", "latest"),

		// these are used when the image stream does not exist and created for the first time by the operator
		MDCImageStreamInitialImage:        getEnv("MDC_IMAGE_STREAM_INITIAL_IMAGE", "quay.io/aerogear/mobile-developer-console:latest"),
		OauthProxyImageStreamInitialImage: getEnv("OAUTH_PROXY_IMAGE_STREAM_INITIAL_IMAGE", "docker.io/openshift/oauth-proxy:v1.1.0"),

		OAuthClientIdSuffix: getEnv("OAUTH_CLIENT_ID_SUFFIX", "mdc-oauth-client"),
		OAuthClientSecret:   getEnv("OAUTH_CLIENT_SECRET", defaultOAuthClientSecret),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	err := os.Setenv(key, defaultVal)
	if err != nil {
		panic(fmt.Sprintf("Unable to set env var %s with value %s . Error is: %s", key, defaultVal, err))
	}
	return defaultVal
}

func getReqEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	panic("Required env var is missing: " + key)
}

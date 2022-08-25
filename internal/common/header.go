package common

import (
	"encoding/json"
	"os"
)

var (
	HeadersEnvVar = "PHC_HEADERS"
)

func HeaderFromEnv() (map[string]string, error) {
	headStr := os.Getenv(HeadersEnvVar)
	envHeaders := map[string]string{}
	if headStr == "" {
		return envHeaders, nil
	}

	err := json.Unmarshal([]byte(headStr), &envHeaders)
	if err != nil {
		return envHeaders, err
	}

	return envHeaders, nil
}

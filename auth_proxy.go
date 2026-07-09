package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"maps"
	"slices"
	"time"
)

var debug = os.Getenv("RCLONE_SERVER_DEBUG") != ""
var logFile *os.File

func debugLog(format string, args ...any) {
	if !debug || logFile == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s %s\n", time.Now().Format("15:04:05"), msg)
	logFile.WriteString(line)
	logFile.Sync()
}

type AuthInput struct {
	User      string `json:"user"`
	Pass      string `json:"pass"`
	PublicKey string `json:"public_key"`
}

type UserAuth struct {
	Pass       string   `json:"pass"`
	PublicKeys []string `json:"public_keys"`
}

type UserConfig struct {
	Auth   UserAuth          `json:"auth"`
	Config map[string]string `json:"config"`
}

func loadAuthData(path string) (map[string]UserConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading auth file: %w", err)
	}
	var authData map[string]UserConfig
	if err := json.Unmarshal(data, &authData); err != nil {
		return nil, fmt.Errorf("parsing auth file: %w", err)
	}
	return authData, nil
}

func lookupUser(input AuthInput, authData map[string]UserConfig) (UserConfig, bool) {
	cfg, ok := authData[input.User]
	return cfg, ok
}

func checkCredentials(input AuthInput, userAuth UserAuth) bool {
	if input.PublicKey != "" {
		return slices.Contains(userAuth.PublicKeys, input.PublicKey)
	}
	if input.Pass != "" {
		return input.Pass == userAuth.Pass
	}
	return false
}

func buildOutput(input AuthInput, userConfig UserConfig) map[string]string {
	output := map[string]string{
		"user":       input.User,
		"pass":       input.Pass,
		"public_key": input.PublicKey,
		"_obscure":   "pass",
		"_root":      "",
	}
	maps.Copy(output, userConfig.Config)
	return output
}

func authenticate(input AuthInput, authData map[string]UserConfig) (map[string]string, bool) {
	userConfig, found := lookupUser(input, authData)
	if !found {
		debugLog("user %q not found in auth data", input.User)
		return nil, false
	}
	if !checkCredentials(input, userConfig.Auth) {
		debugLog("credentials mismatch for user %q (public_key=%q pass_provided=%v)", input.User, input.PublicKey, input.Pass != "")
		return nil, false
	}
	return buildOutput(input, userConfig), true
}

func main() {
	if debug {
		var err error
		logFile, err = os.OpenFile("/tmp/auth_proxy.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("opening log file: %v", err)
		}
		defer logFile.Close()
	}

	inBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		debugLog("error reading stdin: %v", err)
		log.Fatalf("reading stdin: %v", err)
	}

	var input AuthInput
	if err := json.Unmarshal(inBytes, &input); err != nil {
		debugLog("error parsing auth input: %v (raw: %s)", err, string(inBytes))
		log.Fatalf("parsing auth input: %v", err)
	}

	authPath := os.Getenv("AUTH_CONFIG")
	debugLog("auth input: %+v", input)
	debugLog("auth config path: %s", authPath)

	authData, err := loadAuthData(authPath)
	if err != nil {
		debugLog("error loading auth data: %v", err)
		log.Fatalf("loading auth data: %v", err)
	}
	debugLog("loaded %d users from auth data", len(authData))

	output, ok := authenticate(input, authData)
	debugLog("matched: %v", ok)
	if !ok {
		os.Exit(1)
	}

	debugLog("final config: %+v", output)

	bytes, _ := json.Marshal(output)
	os.Stdout.Write(bytes)
}

package main

import (
	_ "embed"
	"encoding/json"
	"io"
	"os"
)

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

func getAuthData() map[string]UserConfig {
	auth_data, _ := os.ReadFile(os.Getenv("AUTH_DATA_FILE"))
	var auth_data_map map[string]UserConfig
	_ = json.Unmarshal([]byte(auth_data), &auth_data_map)
	return auth_data_map
}

func main() {
	in_bytes, _ := io.ReadAll(os.Stdin)
	var auth_input AuthInput
	json.Unmarshal(in_bytes, &auth_input)
	auth_data_map := getAuthData()
	user_config, matched := auth_data_map[auth_input.User]
	if matched {
		if auth_input.PublicKey != "" {
			matched = false
			for _, publicKey := range user_config.Auth.PublicKeys {
				if publicKey == auth_input.PublicKey {
					matched = true
					break
				}
			}
		} else if auth_input.Pass != "" {
			matched = auth_input.Pass == user_config.Auth.Pass
		} else {
			matched = false
		}
	}
	if matched {
		data := map[string]string{}
		data["user"] = auth_input.User
		data["pass"] = auth_input.Pass
		data["public_key"] = auth_input.PublicKey
		data["_obscure"] = "pass"
		data["_root"] = ""
		for k, v := range user_config.Config {
			data[k] = v
		}
		bytes, _ := json.Marshal(data)
		os.Stdout.WriteString(string(bytes))
	} else {
		os.Exit(1)
	}
}

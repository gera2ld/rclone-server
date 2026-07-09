package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLookupUser(t *testing.T) {
	authData := map[string]UserConfig{
		"alice": {
			Auth:   UserAuth{Pass: "secret", PublicKeys: []string{"ssh-ed25519 AAAA..."}},
			Config: map[string]string{"remote": "local"},
		},
	}

	t.Run("found", func(t *testing.T) {
		input := AuthInput{User: "alice"}
		cfg, ok := lookupUser(input, authData)
		if !ok {
			t.Fatal("expected user to be found")
		}
		if cfg.Auth.Pass != "secret" {
			t.Errorf("got pass %q, want %q", cfg.Auth.Pass, "secret")
		}
	})

	t.Run("not found", func(t *testing.T) {
		input := AuthInput{User: "bob"}
		_, ok := lookupUser(input, authData)
		if ok {
			t.Fatal("expected user not found")
		}
	})
}

func TestCheckCredentials(t *testing.T) {
	userAuth := UserAuth{
		Pass:       "s3cret",
		PublicKeys: []string{"ssh-ed25519 AAAA...", "ssh-rsa BBBB..."},
	}

	t.Run("password match", func(t *testing.T) {
		input := AuthInput{Pass: "s3cret"}
		if !checkCredentials(input, userAuth) {
			t.Error("expected password to match")
		}
	})

	t.Run("password mismatch", func(t *testing.T) {
		input := AuthInput{Pass: "wrong"}
		if checkCredentials(input, userAuth) {
			t.Error("expected password not to match")
		}
	})

	t.Run("public key match", func(t *testing.T) {
		input := AuthInput{PublicKey: "ssh-ed25519 AAAA..."}
		if !checkCredentials(input, userAuth) {
			t.Error("expected public key to match")
		}
	})

	t.Run("public key mismatch", func(t *testing.T) {
		input := AuthInput{PublicKey: "ssh-ed25519 unknown"}
		if checkCredentials(input, userAuth) {
			t.Error("expected public key not to match")
		}
	})

	t.Run("no credentials provided", func(t *testing.T) {
		input := AuthInput{}
		if checkCredentials(input, userAuth) {
			t.Error("expected no credentials to fail")
		}
	})

	t.Run("public key takes precedence over password", func(t *testing.T) {
		input := AuthInput{PublicKey: "ssh-ed25519 AAAA...", Pass: "wrong"}
		if !checkCredentials(input, userAuth) {
			t.Error("expected public key match even with wrong password")
		}
	})

	t.Run("public key mismatch ignores password", func(t *testing.T) {
		input := AuthInput{PublicKey: "wrong", Pass: "s3cret"}
		if checkCredentials(input, userAuth) {
			t.Error("expected wrong public key to fail even with correct password")
		}
	})
}

func TestBuildOutput(t *testing.T) {
	input := AuthInput{User: "alice", Pass: "mypass", PublicKey: "ssh-ed25519 AAAA..."}
	userConfig := UserConfig{
		Config: map[string]string{
			"remote":  "local",
			"_root":   "/data",
			"_obscure": "pass",
		},
	}

	output := buildOutput(input, userConfig)

	// Fixed keys
	if output["user"] != "alice" {
		t.Errorf("user: got %q, want %q", output["user"], "alice")
	}
	if output["pass"] != "mypass" {
		t.Errorf("pass: got %q, want %q", output["pass"], "mypass")
	}
	if output["public_key"] != "ssh-ed25519 AAAA..." {
		t.Errorf("public_key: got %q", output["public_key"])
	}
	if output["_obscure"] != "pass" {
		t.Errorf("_obscure: got %q, want %q", output["_obscure"], "pass")
	}

	// Config merge
	if output["remote"] != "local" {
		t.Errorf("remote: got %q, want %q", output["remote"], "local")
	}
}

func TestBuildOutput_emptyConfig(t *testing.T) {
	input := AuthInput{User: "alice", Pass: "pass"}
	userConfig := UserConfig{}

	output := buildOutput(input, userConfig)

	if len(output) != 5 {
		t.Errorf("expected 5 keys, got %d: %v", len(output), output)
	}
	if output["user"] != "alice" {
		t.Errorf("user: got %q, want %q", output["user"], "alice")
	}
}

func TestAuthenticate(t *testing.T) {
	authData := map[string]UserConfig{
		"alice": {
			Auth:   UserAuth{Pass: "secret", PublicKeys: []string{"ssh-ed25519 AAAA..."}},
			Config: map[string]string{"remote": "local"},
		},
		"bob": {
			Auth:   UserAuth{Pass: "bobpass"},
			Config: map[string]string{"remote": "s3"},
		},
	}

	t.Run("password auth success", func(t *testing.T) {
		input := AuthInput{User: "alice", Pass: "secret"}
		output, ok := authenticate(input, authData)
		if !ok {
			t.Fatal("expected authentication to succeed")
		}
		if output["user"] != "alice" {
			t.Errorf("user: got %q", output["user"])
		}
		if output["remote"] != "local" {
			t.Errorf("remote: got %q, want %q", output["remote"], "local")
		}
	})

	t.Run("public key auth success", func(t *testing.T) {
		input := AuthInput{User: "alice", PublicKey: "ssh-ed25519 AAAA..."}
		output, ok := authenticate(input, authData)
		if !ok {
			t.Fatal("expected authentication to succeed")
		}
		if output["public_key"] != "ssh-ed25519 AAAA..." {
			t.Errorf("public_key: got %q", output["public_key"])
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		input := AuthInput{User: "alice", Pass: "wrong"}
		_, ok := authenticate(input, authData)
		if ok {
			t.Fatal("expected authentication to fail")
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		input := AuthInput{User: "unknown", Pass: "pass"}
		_, ok := authenticate(input, authData)
		if ok {
			t.Fatal("expected authentication to fail")
		}
	})

	t.Run("no credentials", func(t *testing.T) {
		input := AuthInput{User: "alice"}
		_, ok := authenticate(input, authData)
		if ok {
			t.Fatal("expected authentication to fail")
		}
	})

	t.Run("bob password auth", func(t *testing.T) {
		input := AuthInput{User: "bob", Pass: "bobpass"}
		output, ok := authenticate(input, authData)
		if !ok {
			t.Fatal("expected authentication to succeed")
		}
		if output["remote"] != "s3" {
			t.Errorf("remote: got %q, want %q", output["remote"], "s3")
		}
	})
}

func TestLoadAuthData(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "auth.json")
		data := map[string]UserConfig{
			"alice": {
				Auth:   UserAuth{Pass: "secret"},
				Config: map[string]string{"remote": "local"},
			},
		}
		raw, _ := json.Marshal(data)
		os.WriteFile(path, raw, 0644)

		authData, err := loadAuthData(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if authData["alice"].Auth.Pass != "secret" {
			t.Errorf("got pass %q, want %q", authData["alice"].Auth.Pass, "secret")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loadAuthData("/nonexistent/auth.json")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "auth.json")
		os.WriteFile(path, []byte("not json"), 0644)

		_, err := loadAuthData(path)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "auth.json")
		os.WriteFile(path, []byte("{}"), 0644)

		authData, err := loadAuthData(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(authData) != 0 {
			t.Errorf("expected empty map, got %d entries", len(authData))
		}
	})
}

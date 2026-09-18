package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesEnvironmentAfterFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{
		"server":{"host":"127.0.0.1","port":8080},
		"ai":{"api_url":"https://example.invalid","api_key":"","model":"file-model"},
		"storage":{"data_dir":"./file-data","auto_save":true,"save_interval":30}
	}`), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BILIBILI_PORT", "9090")
	t.Setenv("BILIBILI_DATA_DIR", "./env-data")
	t.Setenv("ZHIPU_MODEL", "env-model")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 9090 || cfg.Storage.DataDir != "./env-data" || cfg.AI.Model != "env-model" {
		t.Fatalf("environment did not override file: %#v", cfg)
	}
}

func TestLoadRejectsMalformedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"server":`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected malformed config to fail")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("BILIBILI_PORT", "70000")
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected invalid port to fail validation")
	}
}

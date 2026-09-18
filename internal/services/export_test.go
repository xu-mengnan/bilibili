package services

import "testing"

func TestValidateExportBasename(t *testing.T) {
	valid := []string{"comments", "评论导出", "my-report_2026", "a.b"}
	for _, name := range valid {
		if _, err := validateExportBasename(name); err != nil {
			t.Fatalf("expected %q to be valid: %v", name, err)
		}
	}

	invalid := []string{
		"../secret",
		"..\\secret",
		"/tmp/secret",
		"C:\\temp\\secret",
		"a/b",
		"a\\b",
		"evil<script>",
	}
	for _, name := range invalid {
		if _, err := validateExportBasename(name); err == nil {
			t.Fatalf("expected %q to be rejected", name)
		}
	}
}

func TestNormalizeExportFormat(t *testing.T) {
	cases := map[string]string{
		"xlsx":  "xlsx",
		"excel": "xlsx",
		"CSV":   "csv",
	}
	for input, want := range cases {
		got, err := normalizeExportFormat(input)
		if err != nil {
			t.Fatalf("normalize %q: %v", input, err)
		}
		if got != want {
			t.Fatalf("normalize %q = %q, want %q", input, got, want)
		}
	}

	if _, err := normalizeExportFormat("html"); err == nil {
		t.Fatal("expected unsupported format to fail")
	}
}

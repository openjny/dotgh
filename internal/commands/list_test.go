package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestTemplatesDir creates a temporary templates directory with the given template names.
// Returns the path to the templates directory.
func setupTestTemplatesDir(t *testing.T, templates []string) string {
	t.Helper()
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "dotgh", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates directory: %v", err)
	}
	for _, tmpl := range templates {
		if err := os.MkdirAll(filepath.Join(templatesDir, tmpl), 0755); err != nil {
			t.Fatalf("failed to create template %s: %v", tmpl, err)
		}
	}
	return templatesDir
}

// executeListCmd runs the list command with the given templates directory and returns the output.
func executeListCmd(t *testing.T, templatesDir string) (string, error) {
	t.Helper()
	cmd := NewListCmd(templatesDir)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.Execute()
	return buf.String(), err
}

func TestRunList(t *testing.T) {
	tests := []struct {
		name           string
		setupTemplates []string // テンプレートディレクトリ名のリスト
		wantContains   []string // 出力に含まれるべき文字列
		wantErr        bool
	}{
		{
			name:           "no templates",
			setupTemplates: []string{},
			wantContains:   []string{"Available templates:", "(no templates found)"},
			wantErr:        false,
		},
		{
			name:           "single template",
			setupTemplates: []string{"my-template"},
			wantContains:   []string{"Available templates:", "my-template", "1 template(s) found"},
			wantErr:        false,
		},
		{
			name:           "multiple templates",
			setupTemplates: []string{"template-a", "template-b", "template-c"},
			wantContains:   []string{"Available templates:", "template-a", "template-b", "template-c", "3 template(s) found"},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templatesDir := setupTestTemplatesDir(t, tt.setupTemplates)
			output, err := executeListCmd(t, templatesDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("runList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

func TestRunListWithNonExistentDir(t *testing.T) {
	// 存在しないディレクトリを指定
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "non-existent", "templates")

	output, err := executeListCmd(t, nonExistentDir)
	if err != nil {
		t.Errorf("runList() should not return error for non-existent dir, got: %v", err)
	}

	if !strings.Contains(output, "(no templates found)") {
		t.Errorf("output should indicate no templates found, got:\n%s", output)
	}
}

func TestRunListIgnoresFiles(t *testing.T) {
	// ファイルはテンプレートとして扱わないことを確認
	templatesDir := setupTestTemplatesDir(t, []string{"real-template"})

	// ファイルを1つ作成（これは無視されるべき）
	if err := os.WriteFile(filepath.Join(templatesDir, "not-a-template.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	output, err := executeListCmd(t, templatesDir)
	if err != nil {
		t.Errorf("runList() error = %v", err)
		return
	}

	if !strings.Contains(output, "real-template") {
		t.Errorf("output should contain 'real-template', got:\n%s", output)
	}
	if strings.Contains(output, "not-a-template") {
		t.Errorf("output should NOT contain 'not-a-template', got:\n%s", output)
	}
	if !strings.Contains(output, "1 template(s) found") {
		t.Errorf("output should show '1 template(s) found', got:\n%s", output)
	}
}

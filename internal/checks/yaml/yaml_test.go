// internal/checks/yaml/yaml_test.go
package yaml_test

import (
	"testing"

	"aint/internal/checks/yaml"
)

func TestConditionalRunnerSizingDetectsTernary(t *testing.T) {
	line := `    runs-on: ${{ github.event_name == 'pull_request' && 'c7i.8xlarge' || 'c7i.24xlarge' }}`
	findings := yaml.ConditionalRunnerSizing.Run("build.yml", []byte(line), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestConditionalRunnerSizingIgnoresFixedRunner(t *testing.T) {
	line := `    runs-on: c7i.24xlarge`
	findings := yaml.ConditionalRunnerSizing.Run("build.yml", []byte(line), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

package sarif

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

var update = flag.Bool("update", false, "update golden SARIF files")

// TestRenderConformsToSARIFSchema validates that the emitted document conforms to
// the official SARIF 2.1.0 JSON schema (vendored under testdata/), so correctness
// is pinned to the real spec rather than only our own golden expectation.
func TestRenderConformsToSARIFSchema(t *testing.T) {
	data, err := Render(sampleResult())
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	schemaBytes, err := os.ReadFile(filepath.Join("testdata", "sarif-2.1.0.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	schemaDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource("sarif-2.1.0.schema.json", schemaDoc); err != nil {
		t.Fatalf("add schema resource: %v", err)
	}
	sch, err := c.Compile("sarif-2.1.0.schema.json")
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}

	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unmarshal instance: %v", err)
	}
	if err := sch.Validate(inst); err != nil {
		t.Fatalf("SARIF output does not conform to sarif-2.1.0:\n%v", err)
	}
}

// TestRenderGolden locks the exact rendered bytes for a fixed result so any
// mapping change is a visible, intentional diff. Regenerate with -update.
func TestRenderGolden(t *testing.T) {
	data, err := Render(sampleResult())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		t.Fatalf("indent: %v", err)
	}
	got := append(pretty.Bytes(), '\n')

	golden := filepath.Join("testdata", "golden.sarif.json")
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run `go test -run TestRenderGolden -update` to create): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("SARIF golden drift; re-run with -update if intended.\n--- got ---\n%s", got)
	}
}

package script

import (
	"os"
	"testing"
)

func TestScriptEngine_Builtins(t *testing.T) {
	engine := NewScriptEngine()

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, res interface{})
	}{
		{
			name: "path.join",
			script: `path.join("a", "b", "c")`,
			check: func(t *testing.T, res interface{}) {
				expected := "a" + string(os.PathSeparator) + "b" + string(os.PathSeparator) + "c"
				if res.(string) != expected {
					t.Errorf("expected %s, got %v", expected, res)
				}
			},
		},
		{
			name: "base64",
			script: `
				var encoded = base64.encode("hello");
				var decoded = base64.decode(encoded);
				decoded;
			`,
			check: func(t *testing.T, res interface{}) {
				if res.(string) != "hello" {
					t.Errorf("expected hello, got %v", res)
				}
			},
		},
		{
			name: "json",
			script: `
				var obj = {a: 1, b: "test"};
				var str = json.stringify(obj);
				var parsed = json.parse(str);
				parsed.b;
			`,
			check: func(t *testing.T, res interface{}) {
				if res.(string) != "test" {
					t.Errorf("expected test, got %v", res)
				}
			},
		},
		{
			name: "os.getenv",
			script: `
				os.setenv("TEST_VAR", "123");
				os.getenv("TEST_VAR");
			`,
			check: func(t *testing.T, res interface{}) {
				if res.(string) != "123" {
					t.Errorf("expected 123, got %v", res)
				}
			},
		},
		{
			name: "utils.uuid",
			script: `utils.uuid().length > 0`,
			check: func(t *testing.T, res interface{}) {
				if res.(bool) != true {
					t.Errorf("expected true, got %v", res)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := engine.Run(tt.script)
			if err != nil {
				t.Fatalf("failed to run script: %v", err)
			}
			tt.check(t, val.Export())
		})
	}
}

func TestScriptEngine_FS(t *testing.T) {
	engine := NewScriptEngine()
	tmpFile := "test_script_fs.txt"
	defer os.Remove(tmpFile)

	script := `
		fs.writeFile("test_script_fs.txt", "hello world");
		var content = fs.readFile("test_script_fs.txt");
		content;
	`

	val, err := engine.Run(script)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}

	if val.String() != "hello world" {
		t.Errorf("expected hello world, got %v", val)
	}
	
	// Test exists
	val, err = engine.Run(`fs.exists("test_script_fs.txt")`)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}
	if !val.ToBoolean() {
		t.Errorf("expected file to exist")
	}

	// Test remove
	_, err = engine.Run(`fs.remove("test_script_fs.txt")`)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}
	
	val, err = engine.Run(`fs.exists("test_script_fs.txt")`)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}
	if val.ToBoolean() {
		t.Errorf("expected file to be removed")
	}
}

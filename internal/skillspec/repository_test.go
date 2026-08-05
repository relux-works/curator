package skillspec

import (
	"strings"
	"testing"
)

func schema7Manifest(repository, command string) string {
	return `{"schema_version":7,"capabilities":{},"build_roots":["build"],"build_repositories":{"tools":` + repository + `},"commands":{"local":{"type":"build","driver":"go-v1","source_dir":"build/cmd/local"},"golden":` + command + `}}`
}

func validRepository() string {
	return `{"git":"https://GIT.example.com/Skills/tools.git","locked_commit":{"object_format":"sha1","hex":"0123456789abcdef0123456789abcdef01234567"},"tag":"v1.4.0"}`
}

func validRepositoryCommand() string {
	return `{"type":"build","driver":"go-repository-v1","repository":"tools","target":"golden-tool"}`
}

func TestSchema7RepositoryModelAndCommandOwnership(t *testing.T) {
	files := map[string]string{"build/go.mod": "module local\n", "build/cmd/local/main.go": "package main\n"}
	spec, err := Load(writeSkill(t, schema7Manifest(validRepository(), validRepositoryCommand()), files))
	if err != nil {
		t.Fatal(err)
	}
	repository := spec.BuildRepositories["tools"]
	if repository.Identity != "git.example.com/Skills/tools" || repository.Transport != "https" || repository.Tag != "v1.4.0" {
		t.Fatalf("repository = %+v", repository)
	}
	if repository.LockedCommit.ObjectFormat != "sha1" || len(repository.LockedCommit.Hex) != 40 {
		t.Fatalf("lock = %+v", repository.LockedCommit)
	}
	command := spec.Commands["golden"]
	if command.Name != "golden" || command.Repository != "tools" || command.Target != "golden-tool" || command.SourceDir != "" {
		t.Fatalf("command key did not own executable identity: %+v", command)
	}
}

func TestSchema7RepositoryCommandAndDeclarationStayClosed(t *testing.T) {
	files := map[string]string{"build/go.mod": "module local\n", "build/cmd/local/main.go": "package main\n"}
	for _, field := range []string{"argv", "env", "output", "name", "credentials", "signing", "hooks", "plugins", "generator", "fallback"} {
		command := strings.TrimSuffix(validRepositoryCommand(), "}") + `,"` + field + `":{}}`
		if field == "argv" {
			command = strings.TrimSuffix(validRepositoryCommand(), "}") + `,"argv":[]}`
		}
		if _, err := Load(writeSkill(t, schema7Manifest(validRepository(), command), files)); err == nil {
			t.Errorf("command accepted package-controlled field %q", field)
		}
	}
	for _, field := range []string{"branch", "ref", "credentials", "signing", "output"} {
		repository := strings.TrimSuffix(validRepository(), "}") + `,"` + field + `":"x"}`
		if _, err := Load(writeSkill(t, schema7Manifest(repository, validRepositoryCommand()), files)); err == nil {
			t.Errorf("repository accepted field %q", field)
		}
	}
}

func TestSchema7RepositoryReferencesAreExactAndExhaustive(t *testing.T) {
	files := map[string]string{"build/go.mod": "module local\n", "build/cmd/local/main.go": "package main\n"}
	missing := `{"type":"build","driver":"go-repository-v1","repository":"missing","target":"golden-tool"}`
	if _, err := Load(writeSkill(t, schema7Manifest(validRepository(), missing), files)); err == nil {
		t.Fatal("accepted command selecting an undeclared repository")
	}
	unselected := `{"schema_version":7,"capabilities":{},"build_repositories":{"tools":` + validRepository() + `},"commands":{}}`
	if _, err := Load(writeSkill(t, unselected, nil)); err == nil {
		t.Fatal("accepted an unselected repository declaration")
	}
}

func TestSchema6RejectsSchema7FieldsWithoutChangingGoV1(t *testing.T) {
	files := map[string]string{"build/go.mod": "module local\n", "build/cmd/local/main.go": "package main\n"}
	legacy := `{"schema_version":6,"capabilities":{},"build_roots":["build"],"commands":{"local":{"type":"build","driver":"go-v1","source_dir":"build/cmd/local"}}}`
	spec, err := Load(writeSkill(t, legacy, files))
	if err != nil || spec.Commands["local"].Driver != "go-v1" {
		t.Fatalf("schema-6 go-v1 regression: %+v, %v", spec, err)
	}
	withRepository := strings.TrimSuffix(legacy, "}") + `,"build_repositories":{"tools":` + validRepository() + `}}`
	if _, err := Load(writeSkill(t, withRepository, files)); err == nil {
		t.Fatal("schema 6 accepted schema 7 build_repositories")
	}
	withCommand := `{"schema_version":6,"capabilities":{},"commands":{"golden":` + validRepositoryCommand() + `}}`
	if _, err := Load(writeSkill(t, withCommand, nil)); err == nil {
		t.Fatal("schema 6 accepted go-repository-v1")
	}
}

func TestSchema7OptionalTagAndSHA256Lock(t *testing.T) {
	repository := `{"git":"git@git.example.com:skills/tools.git","locked_commit":{"object_format":"sha256","hex":"` + strings.Repeat("a", 64) + `"}}`
	manifest := `{"schema_version":7,"capabilities":{},"build_repositories":{"tools":` + repository + `},"commands":{"golden":` + validRepositoryCommand() + `}}`
	spec, err := Load(writeSkill(t, manifest, nil))
	if err != nil {
		t.Fatal(err)
	}
	if got := spec.BuildRepositories["tools"]; got.Tag != "" || got.LockedCommit.ObjectFormat != "sha256" || len(got.LockedCommit.Hex) != 64 {
		t.Fatalf("repository = %+v", got)
	}
}

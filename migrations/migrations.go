package migrations

import (
	"bytes"
	"log"
	"os/exec"
)

var (
	out     bytes.Buffer
	strderr bytes.Buffer
)

func Run() error {
	cmd := exec.Command("go", "run", "./tests/migration-test/testdb_migration_test.go")
	cmd.Stdout = &out
	cmd.Stderr = &strderr

	if err := cmd.Run(); err != nil {
		log.Printf("Migration command failed: %v, stderr: %s", err, strderr.String())
		return err
	}
	log.Printf("Migration completed successfully: %s", out.String())
	return nil
}

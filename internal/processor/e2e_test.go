//go:build !short

package processor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testDBUser   = "testuser"
	testDBPass   = "testpass"
	testDBName   = "testdb"
	pgSchemaDiff = "pg-schema-diff"
)

var (
	binaryName             string
	container              *testContainer
	currentPostgresVersion int
)

func init() {
	// Try to load .env from project root (two levels up from internal/processor)
	envPath := filepath.Join("..", "..", ".env")
	_ = godotenv.Load(envPath)

	if runtime.GOOS == "windows" {
		binaryName = "jacques.exe"
	} else {
		binaryName = "jacques"
	}
}

// getPostgresVersions returns the PostgreSQL versions to test from the POSTGRES_VERSIONS env var.
// Format: "18,17,16" or just "18". Defaults to [18] if not set or invalid.
func getPostgresVersions() []int {
	envVal := os.Getenv("POSTGRES_VERSIONS")
	if envVal == "" {
		return []int{18}
	}

	var versions []int
	for _, v := range strings.Split(envVal, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		var ver int
		if _, err := fmt.Sscanf(v, "%d", &ver); err == nil {
			versions = append(versions, ver)
		}
	}

	if len(versions) == 0 {
		return []int{18}
	}
	return versions
}

// testContainer wraps a PostgreSQL testcontainer with methods for pg_dump and connections.
type testContainer struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
	connStr   string
	ctx       context.Context
}

// TestMain handles one-time setup and teardown for all E2E tests.
// It builds the binary and runs tests against all PostgreSQL versions specified in .env
func TestMain(m *testing.M) {
	// Build the binary before running tests
	if err := buildBinary(); err != nil {
		fmt.Printf("Failed to build binary: %v\n", err)
		os.Exit(1)
	}

	// Get all PostgreSQL versions to test
	versions := getPostgresVersions()

	fmt.Printf("Testing PostgreSQL versions: %v\n", versions)

	// Run tests for each version
	var failed bool
	for i, pgVersion := range versions {
		fmt.Printf("\n=== Running tests with PostgreSQL %d (%d/%d) ===\n", pgVersion, i+1, len(versions))

		// Set the version for this iteration
		currentPostgresVersion = pgVersion

		// Start the PostgreSQL container for this version
		var err error
		container, err = setupContainer()
		if err != nil {
			fmt.Printf("Failed to start container (PostgreSQL %d): %v\n", pgVersion, err)
			failed = true
			continue
		}

		// Run tests
		code := m.Run()

		// Teardown
		if container != nil {
			container.close()
		}

		if code != 0 {
			fmt.Printf("Tests failed for PostgreSQL %d\n", pgVersion)
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}
	os.Exit(0)
}

// buildBinary builds the jacques binary from source.
func buildBinary() error {
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	cmd.Dir = filepath.Join("..", "..")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build binary: %w\nOutput: %s", err, output)
	}
	return nil
}

// setupContainer starts a PostgreSQL testcontainer and waits for it to be ready.
func setupContainer() (*testContainer, error) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, fmt.Sprintf("postgres:%d", currentPostgresVersion),
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPass),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Wait for database to be ready
	if err := waitForDatabase(ctx, connStr); err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("database not ready: %w", err)
	}

	// Create a pool for the test container
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	return &testContainer{
		container: pgContainer,
		pool:      pool,
		connStr:   connStr,
		ctx:       ctx,
	}, nil
}

// waitForDatabase polls the database until it's ready to accept connections.
func waitForDatabase(ctx context.Context, connStr string) error {
	maxRetries := 10
	baseDelay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		pool, err := pgxpool.New(ctx, connStr)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				pool.Close()
				return nil
			}
			pool.Close()
		}
		delay := baseDelay * time.Duration(1<<i)
		if delay > 5*time.Second {
			delay = 5 * time.Second
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("database not ready after %d attempts", maxRetries)
}

// close terminates the PostgreSQL container.
func (tc *testContainer) close() {
	if tc != nil {
		if tc.pool != nil {
			tc.pool.Close()
		}
		if tc.container != nil {
			_ = tc.container.Terminate(tc.ctx)
		}
	}
}

// runSQL executes SQL in the container using psql
func (tc *testContainer) runSQL(dbName, sql string) error {
	// Parse connection string to get host, port, user
	connStr := tc.connStr

	// Extract connection details from URL
	// Format: postgresql://user:pass@host:port/dbname?sslmode=...
	connStr = strings.TrimPrefix(connStr, "postgresql://")

	var user, host, port string
	dbname := dbName

	// Split user:pass@host:port
	if atIdx := strings.Index(connStr, "@"); atIdx != -1 {
		userPass := connStr[:atIdx]
		connStr = connStr[atIdx+1:]
		if colonIdx := strings.Index(userPass, ":"); colonIdx != -1 {
			user = userPass[:colonIdx]
		} else {
			user = userPass
		}
	}

	// Split host:port from dbname
	if slashIdx := strings.Index(connStr, "/"); slashIdx != -1 {
		hostPort := connStr[:slashIdx]
		dbPart := connStr[slashIdx+1:]

		if colonIdx := strings.Index(hostPort, ":"); colonIdx != -1 {
			host = hostPort[:colonIdx]
			port = hostPort[colonIdx+1:]
		} else {
			host = hostPort
		}

		// Remove params from dbname
		if qIdx := strings.Index(dbPart, "?"); qIdx != -1 {
			dbname = dbPart[:qIdx]
		}
	}

	if user == "" {
		user = "postgres"
	}
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}

	// Build psql command to run in container
	cmd := []string{"psql", "-U", user, "-h", host, "-p", port, "-d", dbname, "-c", sql}
	_, _, err := tc.container.Exec(tc.ctx, cmd)
	return err
}

// compareSchemas compares two databases using pg-schema-diff.
func compareSchemas(t *testing.T, fromDSN, toDSN, msg string) {
	t.Helper()

	// Check pg-schema-diff is available
	if _, err := exec.LookPath(pgSchemaDiff); err != nil {
		t.Fatalf("pg-schema-diff not found. Install: go install github.com/stripe/pg-schema-diff/cmd/pg-schema-diff@latest")
	}

	diffCmd := exec.Command(pgSchemaDiff, "plan", "--from-dsn", fromDSN, "--to-dsn", toDSN)
	diffOutput, diffErr := diffCmd.CombinedOutput()

	require.NoError(t, diffErr, "%s\n\nOutput:\n%s", msg, string(diffOutput))
}

// replaceDB replaces the database name in a connection string.
func replaceDB(connStr, newDB string) string {
	// URL-style: postgresql://host:port/dbname?params
	for i := len(connStr) - 1; i >= 0; i-- {
		if connStr[i] == '/' {
			for j := i + 1; j < len(connStr); j++ {
				if connStr[j] == '?' {
					return connStr[:i+1] + newDB + connStr[j:]
				}
			}
			return connStr[:i+1] + newDB
		}
	}
	// Space-separated: host=... dbname=...
	for i := len(connStr) - 1; i >= 0; i-- {
		if connStr[i] == ' ' {
			return connStr[:i+1] + "dbname=" + newDB + connStr[i:]
		}
	}
	return connStr + " dbname=" + newDB
}

// createTestDBs creates test databases using the base connection string.
func createTestDBs(t *testing.T, connStrBase string, names ...string) map[string]string {
	t.Helper()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStrBase)
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, pool.Ping(ctx))

	results := make(map[string]string)
	for _, name := range names {
		_, err := pool.Exec(ctx, "DROP DATABASE IF EXISTS "+name)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "CREATE DATABASE "+name)
		require.NoError(t, err)
		results[name] = replaceDB(connStrBase, name)
	}
	return results
}

// cleanupTestDBs drops the test databases.
func cleanupTestDBs(t *testing.T, connStrBase string, names ...string) {
	t.Helper()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStrBase)
	if err != nil {
		return
	}
	defer pool.Close()

	for _, name := range names {
		_, _ = pool.Exec(ctx, "DROP DATABASE IF EXISTS "+name)
	}
}

// runCleaner runs the cleaner tool on the given input and returns the output.
func runCleaner(t *testing.T, input string) string {
	t.Helper()

	// Get absolute path to binary from project root
	cwd, err := os.Getwd()
	require.NoError(t, err)
	binaryPath := filepath.Join(cwd, "..", "..", binaryName)

	cmd := exec.Command(binaryPath)
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.Output()
	require.NoError(t, err, "cleaner failed.\nInput:\n%s\n\nOutput:\n%s", input, string(output))
	return string(output)
}

// TestE2E_ValidateInfrastructure is a quick sanity check that deploying the same
// schema to two databases produces identical schemas. This validates the test
// environment (container, pg-schema-diff, database) works correctly.
func TestE2E_ValidateInfrastructure(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E tests require full test run: go test ./...")
	}

	inputSQL := LoadFixture(t, "testdata/e2e/", "_basic_input.sql")
	ctx := context.Background()

	dbNames := []string{"infra_db1", "infra_db2"}
	connStrs := createTestDBs(t, container.connStr, dbNames...)
	defer cleanupTestDBs(t, container.connStr, dbNames...)

	// Deploy schema to both databases
	for _, name := range dbNames {
		pool, err := pgxpool.New(ctx, connStrs[name])
		require.NoError(t, err, "failed to connect to %s", name)
		defer pool.Close()
		require.NoError(t, pool.Ping(ctx))
		_, err = pool.Exec(ctx, inputSQL)
		require.NoError(t, err, "failed to deploy schema to %s", name)
	}

	// Compare schemas - should be identical
	compareSchemas(t, connStrs[dbNames[0]], connStrs[dbNames[1]],
		"INFRASTRUCTURE BUG: Same SQL should produce identical schemas")
}

// TestE2E_PgDumpCleanRemigrate tests that applying the cleaner to a schema
// produces a semantically equivalent result:
//
// 1. Load a test fixture (input + expected)
// 2. Run cleaner on input
// 3. Compare cleaned output to expected (exact match - validates transformation correctness)
// 4. Load input into "orig" database
// 5. Load cleaned output into "clean" database
// 6. Compare schemas (semantic equivalence - validates pg-schema-diff compatibility)
func TestE2E_PgDumpCleanRemigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E tests require full test run: go test ./...")
	}

	// Dynamically discover all E2E fixtures from testdata/e2e/
	fixtures := DiscoverFixtures("testdata/e2e/")
	require.NotEmpty(t, fixtures, "no E2E fixtures found")

	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			inputSQL := LoadFixture(t, f.Dir, f.Name)
			expectedSQL := LoadFixture(t, f.Dir, strings.Replace(f.Name, "_input.sql", "_expected.sql", 1))

			// Step 2: Run cleaner on fixture
			cleanedOutput := runCleaner(t, inputSQL)

			// Step 3: Compare cleaned output to expected (exact match)
			assert.Equal(t, NormalizeSQL(expectedSQL), NormalizeSQL(cleanedOutput),
				"Cleaned output should match expected (transformation correctness)")

			dbNames := []string{"orig", "clean"}
			connStrs := createTestDBs(t, container.connStr, dbNames...)
			defer cleanupTestDBs(t, container.connStr, dbNames...)

			// Step 4: Load fixture into original database (via psql for metacommands support)
			require.NoError(t, container.runSQL(dbNames[0], inputSQL),
				"failed to load fixture into original database")

			// Step 5: Load cleaned schema into second database (via psql for metacommands support)
			require.NoError(t, container.runSQL(dbNames[1], cleanedOutput),
				"failed to load cleaned schema.\nCleaned SQL:\n%s", cleanedOutput)

			// Step 6: Compare schemas - should be semantically equivalent
			compareSchemas(t, connStrs[dbNames[0]], connStrs[dbNames[1]],
				"Cleaner should produce semantically equivalent schema")
		})
	}
}

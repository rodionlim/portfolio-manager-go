package migrations

// MigratorInterface defines the contract for database migrations
type MigratorInterface interface {
	// Migrate runs all necessary migrations for the current version
	Migrate() error
	// GetLastMigrationApplied returns the version of the last migration that was applied
	GetLastMigrationApplied() string
}

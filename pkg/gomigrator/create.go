package gomigrator

import "github.com/hilltracer/gomigrator/internal/creator"

// Create generates a timestamp-prefixed SQL migration file and
// returns its absolute path.
func Create(dir, name string) (string, error) { return creator.Create(dir, name) }

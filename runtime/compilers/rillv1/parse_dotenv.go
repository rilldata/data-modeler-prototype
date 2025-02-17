package rillv1

import (
	"context"

	"github.com/joho/godotenv"
)

// parseDotEnv parses the .env file at the given path and merges it with the existing env vars
func (p *Parser) parseDotEnv(ctx context.Context, path string) error {
	data, err := p.Repo.Get(ctx, path)
	if err != nil {
		return err
	}
	envMap, err := godotenv.Unmarshal(data)
	if err != nil {
		return err
	}

	if p.DotEnv == nil {
		p.DotEnv = make(DotEnvMap)
	}

	// Store the env vars for this specific path
	p.DotEnv[path] = envMap

	return nil
}

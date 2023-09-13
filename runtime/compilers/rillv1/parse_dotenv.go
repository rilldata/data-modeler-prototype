package rillv1

import (
	"context"
	"os"

	"github.com/joho/godotenv"
)

// parseDotEnv parses the env file present at repo root
func (p *Parser) parseDotEnv(ctx context.Context, path string) error {
	data, err := p.Repo.Get(ctx, path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	envMap, err := godotenv.Unmarshal(data)
	if err != nil {
		return err
	}

	p.DotEnv = envMap
	return nil
}

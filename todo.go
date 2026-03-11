// services/todo/todo.go
package todo

import (
	"context"

	"github.com/pivaldi/mmw/todo/internal/infra/config"
	"github.com/rotisserie/eris"
)

func GetConfig(ctx context.Context, envs map[string]string) (*config.Config, error) {
	conf, err := config.Load(ctx, envs)
	if err != nil {
		return nil, eris.Wrap(err, "failed to load todo configuration")
	}

	return conf, nil
}

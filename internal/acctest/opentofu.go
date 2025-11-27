package acctest

import (
	"os"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/env"
)

func IsRunningOpenTofu() bool {
	return os.Getenv(env.AccRunningOpenTofu) == "true"
}

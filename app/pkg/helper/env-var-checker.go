package helper

import (
	"fmt"
	"os"

	"github.com/rotisserie/eris"
)

type EnvVarChecker struct {
	envVars []string
}

func NewEnvVarChecker(vars ...string) *EnvVarChecker {
	return &EnvVarChecker{
		envVars: vars,
	}
}

func (evc *EnvVarChecker) Check() error {
	for _, key := range evc.envVars {
		_, exists := os.LookupEnv(key)

		if !exists {
			return eris.New(fmt.Sprintf("Env var %s not set", key))
		}
	}

	return nil
}

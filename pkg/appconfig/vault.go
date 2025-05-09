package appconfig

import (
	"context"
	"errors"
	"github.com/3lvia/libraries-go/pkg/hashivault"
	log "github.com/sirupsen/logrus"
	logDefault "log"
	"os"
)

func setupVaultClient(ctx context.Context, environment string) (hashivault.SecretsManager, error) {
	vaultAddress := func() string {
		if environment == "prod" {
			return "https://vault.elvia.io"
		}

		if environment == "sandbox" || environment == "dev" {
			return "https://vault.dev-elvia.io"
		}

		return "https://vault." + environment + "-elvia.io"
	}()

	vaultClient, errChan, err := hashivault.New(
		ctx,
		hashivault.WithOIDC(),
		hashivault.WithVaultToken(os.Getenv("VAULT_TOKEN")),
		hashivault.WithVaultAddress(vaultAddress),
		hashivault.WithLogger(logDefault.Default()),
	)
	if err != nil {
		log.Errorf("Could not create Vault client: %+v", err)

		return nil, errors.New("CouldNotCreateVaultClient")
	}

	go func(ec <-chan error) {
		for err := range ec {
			log.Errorf("Vault client error: %+v", err)
		}
	}(errChan)

	return vaultClient, nil
}

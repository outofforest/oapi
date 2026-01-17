package oapi

import (
	"context"
	"os"
	"path/filepath"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
	"github.com/pkg/errors"

	"github.com/outofforest/run"
)

// ImportMapping defines map for imports.
type ImportMapping map[string]string

// Configuration is the config for generator.
type Configuration struct {
	SpecFile       string
	OutputFile     string
	ImportMapping  ImportMapping
	GenerateServer bool
	GenerateClient bool
	GenerateModels bool
}

// Generate generates code from OpenAPI spec file.
func Generate(configs ...Configuration) {
	run.New().Run(context.Background(), "generator", func(ctx context.Context) error {
		for _, cfg := range configs {
			if err := generate(cfg); err != nil {
				return err
			}
		}
		return nil
	})
}

func generate(cfg Configuration) error {
	absDir, err := filepath.Abs(filepath.Dir(cfg.OutputFile))
	if err != nil {
		return err
	}

	opts := codegen.Configuration{
		PackageName: filepath.Base(absDir),
		Generate: codegen.GenerateOptions{
			Strict:     true,
			EchoServer: cfg.GenerateServer,
			Client:     cfg.GenerateClient,
			Models:     cfg.GenerateModels,
		},
		Compatibility: codegen.CompatibilityOptions{
			DisableFlattenAdditionalProperties: true,
			DisableRequiredReadOnlyAsPointer:   true,
			AlwaysPrefixEnumValues:             true,
		},
		OutputOptions: codegen.OutputOptions{
			SkipPrune: true,
		},
		ImportMapping: cfg.ImportMapping,
	}
	if err := opts.Validate(); err != nil {
		return errors.WithStack(err)
	}

	swagger, err := util.LoadSwagger(cfg.SpecFile)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := os.MkdirAll(absDir, 0o700); err != nil {
		return errors.WithStack(err)
	}
	code, err := codegen.Generate(swagger, opts)
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(os.WriteFile(cfg.OutputFile, []byte(code), 0o600))
}

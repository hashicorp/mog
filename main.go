// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR", err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	flags, opts := setupFlags(args[0])
	err := flags.Parse(args[1:])
	switch {
	case err == flag.ErrHelp:
		return nil
	case err != nil:
		return err
	}

	log.SetFlags(0)
	return runMog(*opts)
}

type options struct {
	source                  string
	ignorePackageLoadErrors bool
	tags                    string
}

func (o options) handlePackageLoadErrors(pkg *packages.Package) error {
	if o.ignorePackageLoadErrors {
		// TODO: setup logger
		for _, err := range pkg.Errors {
			log.Println(err.Error())
		}
		return nil
	}
	return packageLoadErrors(pkg)
}

func setupFlags(name string) (*flag.FlagSet, *options) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	opts := &options{}

	// TODO: make this a positional arg, set a Usage func to document it
	flags.StringVar(&opts.source, "source", ".", "package path for source structs")

	// TODO: make this a positional arg, set a Usage func to document it
	flags.StringVar(&opts.tags, "tags", ".", "build tags to be passed when parsing the packages")

	flags.BoolVar(&opts.ignorePackageLoadErrors, "ignore-package-load-errors", false,
		"ignore any syntax errors encountered while loading source")
	return flags, opts
}

func runMog(opts options) error {
	if opts.source == "" {
		return fmt.Errorf("missing required source package")
	}

	sources, err := loadSourceStructs(opts.source, opts.tags, opts.handlePackageLoadErrors)
	if err != nil {
		return fmt.Errorf("failed to load source from %s: %w", opts.source, err)
	}

	cfg, err := configsFromAnnotations(sources)
	if err != nil {
		return fmt.Errorf("failed to parse annotations: %w", err)
	}

	if len(cfg.Structs) == 0 {
		log.Printf("no source structs found in %v", opts.source)
		return nil
	}

	targets, err := loadTargetStructs(targetPackages(cfg.Structs), opts.tags)
	if err != nil {
		return fmt.Errorf("failed to load targets: %w", err)
	}

	cfg.Structs = applyAutoConvertFunctions(cfg.SourcePkg.pkg.PkgPath, cfg.Structs)

	log.Printf("Generating code for %d structs", len(cfg.Structs))

	return generateFiles(cfg, targets)
}

func targetPackages(cfgs []structConfig) []string {
	result := make([]string, 0, len(cfgs))
	for _, cfg := range cfgs {
		if cfg.Target.Package == "" {
			continue
		}
		result = append(result, cfg.Target.Package)
	}
	return result
}

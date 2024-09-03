/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path"

	"k8s.io/klog/v2"
	"sigs.k8s.io/container-object-storage-interface-api/test/generator/config"
	"sigs.k8s.io/container-object-storage-interface-api/test/generator/generator"
)

func main() {
	var (
		rootPath   string
		configPath string
	)
	flag.StringVar(&rootPath, "root", ".", "Root path where the tests will be written")
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	klog.InitFlags(nil)
	flag.Parse()

	if err := run(
		context.Background(),
		rootPath,
		configPath,
	); err != nil {
		log.Fatalf("Generator failed: %s", err)
	}
}

func run(_ context.Context, rootPath, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	tests, err := generator.Matrix(cfg)
	if err != nil {
		return err
	}

	for _, test := range tests {
		if err := os.MkdirAll(
			path.Join(rootPath, "e2e", test.Name, "resources"),
			0o755,
		); err != nil {
			return err
		}

		if err := os.WriteFile(
			path.Join(rootPath, "e2e", test.Name, "chainsaw-test.yaml"),
			test.ChainsawTestSpec,
			0o644,
		); err != nil {
			return err
		}

		if err := os.WriteFile(
			path.Join(rootPath, "e2e", test.Name, "resources", "BucketAccess.yaml"),
			test.Resources.BucketAccess,
			0o644,
		); err != nil {
			return err
		}

		if err := os.WriteFile(
			path.Join(rootPath, "e2e", test.Name, "resources", "BucketAccessClass.yaml"),
			test.Resources.BucketAccessClass,
			0o644,
		); err != nil {
			return err
		}

		if err := os.WriteFile(
			path.Join(rootPath, "e2e", test.Name, "resources", "BucketClaim.yaml"),
			test.Resources.BucketClaim,
			0o644,
		); err != nil {
			return err
		}

		if err := os.WriteFile(
			path.Join(rootPath, "e2e", test.Name, "resources", "BucketClass.yaml"),
			test.Resources.BucketClass,
			0o644,
		); err != nil {
			return err
		}
	}

	return nil
}

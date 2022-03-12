// 🐇 tsubasa: Microservice to define a schema and execute it in a fast environment.
// Copyright 2022 Noel <cutie@floofy.dev>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tsubasa

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

func newGenerateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generates a `config.toml` file in the working directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if len(args) == 1 {
				cwd = args[0]
			} else {
				if err != nil {
					return err
				}
			}

			fmt.Printf("> Now creating `config.toml` in directory '%s'...\n", cwd)
			defaultConfig := `
debug = false

[elastic]
nodes = ["http://localhost:9200"]
`

			file := fmt.Sprintf("%s/config.toml", cwd)
			if err := ioutil.WriteFile(file, []byte(defaultConfig), 0o666); err != nil {
				return err
			} else {
				fmt.Printf("> Created `config.toml` in path '%s'!", file)
				return nil
			}
		},
	}
}

/*
Copyright © 2022 Oleksandr Yershov <oleksandr.yershov@hpe.com>
*/

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	yaml_v3 "gopkg.in/yaml.v3"
	"sigs.k8s.io/yaml"

	"github.com/alex60217101990/json_schema_generator/internal/enums"
	"github.com/alex60217101990/json_schema_generator/internal/parser"
	"github.com/alex60217101990/json_schema_generator/internal/utils"
)

const (
	defaultSchemaDir = "tmp/values.schema.json"
)

var (
	log          zerolog.Logger
	logLevelMode enums.LogLevelMode

	valuesYamlPath string
	schemaPath     string

	version = "0.0.1"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: version,
	Short:   fmt.Sprintf("%s - %s", color.GreenString("schema-generator"), "CLI util for generate values.schema.json file from values.yaml helm chart file"),
	Long:    fmt.Sprintf(`%s - generate json schema from input helm chart values yaml/yml file`, color.GreenString("schema-generator")),
	Run: func(cmd *cobra.Command, _ []string) {
		rootDir := utils.RootDir()

		logLevel := cmd.PersistentFlags().Lookup("level").Value.String()

		err := utils.ChangeLevel(logLevel, &log)
		if err != nil {
			log.Fatal().Stack().Err(errors.WithStack(err)).Msg("")
		}

		defer func() {
			if r := recover(); r != nil {
				log.Fatal().Stack().Err(errors.WithStack(errors.Wrap(err, "panic error"))).Msg("")
			}
		}()

		ok, err := utils.IsInputFromPipe()
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Msg("")
			return
		}

		// read helm chart yaml values file and convert to json...
		var val []byte
		if ok {
			val, err = ioutil.ReadAll(os.Stdin)
		} else {
			val, err = os.ReadFile(valuesYamlPath)
		}

		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Msg("")
			return
		}

		var data yaml_v3.Node
		err = yaml_v3.Unmarshal(val, &data)
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Msg("")
			return
		}

		var bts []byte
		p := &parser.Parser{}
		bts, err = p.Init(&log).ParseSync(&data, val)
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Msg("")
			return
		}

		var valuesJSON []byte
		valuesJSON, err = yaml.YAMLToJSON(val)
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Msg("")
			return
		}

		if bytes.Equal(valuesJSON, []byte("null")) {
			//nolint:ineffassign,staticcheck
			valuesJSON = []byte("{}")
		}

		if schemaPath == defaultSchemaDir {
			tmpPath := filepath.Join(rootDir, filepath.Dir(defaultSchemaDir))
			if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
				if err = os.MkdirAll(tmpPath, 0755); err != nil {
					log.Error().Stack().Err(errors.WithStack(err)).Msg("")
					return
				}

				defer func() {
					_ = os.RemoveAll(tmpPath)
				}()
			}

			schemaPath = filepath.Join(tmpPath, "values.schema.json")
		}

		log.Info().Msgf("write schema into file: %s", schemaPath)
		err = ioutil.WriteFile(schemaPath, bts, 0644)
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Msg("")
			return
		}

		// generate json schema from input json values...
		info, err := os.Stat(schemaPath)
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Msg("")
			return
		}

		log.Debug().Msgf("json schema: %s file size: %s", schemaPath, utils.HumanSize(float64(info.Size())))
	},
	Use: "schema-generator",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	log = utils.InitLogger(true)

	go utils.OsSignalHandler(nil, nil, log)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Stack().Err(errors.WithMessage(err, "Generator CLI finished with error"))
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().VarP(
		enumflag.New(&logLevelMode, "logLevel", enums.LogLevelModeIds, enumflag.EnumCaseInsensitive),
		"level", "l",
		fmt.Sprintf("Logs level value. can be %s or %s etc.", color.YellowString("'info'"), color.YellowString("'warn'")))
	rootCmd.Flags().StringVarP(&valuesYamlPath, "valuesPath", "v", "", "Path to yaml/yml file with chart values")
	rootCmd.Flags().StringVarP(&schemaPath, "schemaPath", "s", defaultSchemaDir, "Path for json schema file")
}

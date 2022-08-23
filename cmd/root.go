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
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	yaml_v3 "gopkg.in/yaml.v3"
	"sigs.k8s.io/yaml"

	"github.com/alex60217101990/json_schema_generator/internal/parser"
	utils "github.com/alex60217101990/json_schema_generator/internal/utils"
)

const (
	defaultSchemaDir = "tmp/values.schema.json"
)

var (
	log      zerolog.Logger
	logLevel uint8

	valuesYamlPath string
	schemaPath     string

	version = "0.0.1"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Aliases: []string{color.GreenString("sgen")},
	Version: version,
	Short:   fmt.Sprintf("%s - %s", color.GreenString("schema-generator"), "CLI util for generate values.schema.json file from values.yaml helm chart file"),
	Long:    fmt.Sprintf(`%s - generate json schema from input helm chart values yaml/yml file`, color.GreenString("schema-generator")),
	Run: func(cmd *cobra.Command, _ []string) {
		rootDir := utils.RootDir()

		fmt.Printf("log level: %d\n", logLevel)
		// TODO: adding stack tracing for errors.
		utils.ChangeLevel(logLevel, &log)

		defer func() {
			if r := recover(); r != nil {
				log.Fatal().Msgf("panic error: %v", r)
			}
		}()

		// read helm chart yaml values file and convert to json...
		val, err := os.ReadFile(valuesYamlPath)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		var data yaml_v3.Node
		err = yaml_v3.Unmarshal(val, &data)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		//var modified []byte
		var bts []byte
		p := &parser.Parser{}
		bts, err = p.Init(&log).ParseSync(&data, val)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		valuesJSON, err := yaml.YAMLToJSON(val)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		if bytes.Equal(valuesJSON, []byte("null")) {
			valuesJSON = []byte("{}")
		}

		if schemaPath == defaultSchemaDir {
			tmpPath := filepath.Join(rootDir, filepath.Dir(defaultSchemaDir))
			if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
				if err = os.MkdirAll(tmpPath, 0755); err != nil {
					log.Error().Msg(err.Error())
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
			log.Error().Msg(err.Error())
			return
		}

		// generate json schema from input json values...
		info, err := os.Stat(schemaPath)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		log.Debug().Msgf("json schema: %s file size: %s", schemaPath, utils.HumanSize(float64(info.Size())))
	},
	Use: color.GreenString("schema-generator"),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	log = utils.InitLogger()

	go utils.OsSignalHandler(nil, nil, log)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Msgf("Generator CLI finished with error: '%s'", err)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().Uint8VarP(&logLevel, "level", "l", 1, color.BlueString("logs level"))
	rootCmd.Flags().StringVarP(&valuesYamlPath, "valuesPath", "v", "", color.BlueString("Path to yaml/yml file with chart values"))
	rootCmd.Flags().StringVarP(&schemaPath, "schemaPath", "s", defaultSchemaDir, color.BlueString("Path for json schema file"))
}

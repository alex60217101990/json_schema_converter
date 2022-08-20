/*
Copyright © 2022 Oleksandr Yershov <oleksandr.yershov@hpe.com>
*/

package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ccs-installer/uber-installer/src/dind-pipeline-installer/schema-generator/internal/enums"
	"github.com/ccs-installer/uber-installer/src/dind-pipeline-installer/schema-generator/internal/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/fatih/color"
	"github.com/karuppiah7890/go-jsonschema-generator"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	yaml_v3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"

	utils "github.com/ccs-installer/uber-installer/src/dind-pipeline-installer/schema-generator/internal/utils"
)

const (
	defaultSchemaDir          = "tmp/values.schema.json"
	defaultOverrideSchemaPath = "schemas/override-values.schema.json"
)

var (
	log zerolog.Logger

	overrideSchemaPath string
	valuesYamlPath     string
	schemaPath         string

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

		var (
			patchJSON []byte
		)

		tmpDir := os.TempDir()
		defer func() {
			if r := recover(); r != nil {
				_ = os.RemoveAll(tmpDir)
				log.Fatal().Msgf("panic error: %v", r)
			}

			_ = os.RemoveAll(tmpDir)
		}()

		tmpFile, err := ioutil.TempFile(tmpDir, "json-schema-generator-")
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		defer func() {
			if r := recover(); r != nil {
				_ = tmpFile.Close()
				_ = os.Remove(tmpFile.Name())
				_ = os.RemoveAll(tmpDir)
				log.Fatal().Msgf("panic error: %v", r)
			}

			_ = os.Remove(tmpFile.Name())
			_ = os.RemoveAll(tmpDir)
		}()

		// get override values json...
		if overrideSchemaPath == defaultOverrideSchemaPath {
			patchJSON, err = ioutil.ReadFile(filepath.Join(rootDir, overrideSchemaPath))
		} else {
			patchJSON, err = ioutil.ReadFile(overrideSchemaPath)
		}

		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

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

		p := &utils.Parser{}
		p.Init(&log).Start(&data, "", false)

		valuesJSON, err := yaml.YAMLToJSON(val)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		if bytes.Equal(valuesJSON, []byte("null")) {
			valuesJSON = []byte("{}")
		}

		var v chartutil.Values
		err = yaml_v3.Unmarshal(val, &v)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		schema := &jsonschema.Document{}
		schema.ReadDeep(&v)
		jsonBts, err := schema.Marshal()
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		var modified []byte

		patchesJSON := []*types.Patch{{
			OperationType: enums.Replace,
			Path:          "/$schema",
			Value:         []byte(`"https://json-schema.org/draft/2019-09/schema"`),
		}}

		for path, object := range p.ChangeAllPath(jsonBts) {
			patchesJSON = append(patchesJSON, &types.Patch{
				OperationType: enums.Replace,
				Path:          path,
				Value:         []byte(object),
			})
		}

		tmpBts, err := json.Marshal(patchesJSON)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		patch, err := jsonpatch.DecodePatch(tmpBts)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		modified, err = patch.Apply(jsonBts)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		bts, err := utils.PrettyString(modified)
		if err != nil {
			log.Error().Msg(err.Error())
			return
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

			newSchema := filepath.Join(tmpPath, "new-values.schema.json")
			err = ioutil.WriteFile(newSchema, bts, 0644)
			if err != nil {
				log.Error().Msg(err.Error())
				return
			}

			schemaPath = filepath.Join(tmpPath, "values.schema.json")
		}

		bts, err = json.MarshalIndent(valuesJSON, "", "\t")
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		bts, err = base64.StdEncoding.DecodeString(string(bts[1 : len(bts)-1]))
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		err = ioutil.WriteFile(tmpFile.Name(), bts, 0644)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		log.Info().Msgf("write json from values file into %s file", tmpFile.Name())

		// generate json schema from input json values...
		info, err := os.Stat(tmpFile.Name())
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		log.Info().Msgf("json input file size: %s", utils.HumanSize(float64(info.Size())))

		cmdBash := exec.Command(`genson`, tmpFile.Name())
		stdout, err := cmdBash.Output()

		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		if len(stdout) > 0 {
			log.Info().Msgf("generate json schema with %s success", color.GreenString("genson"))
		}

		// merge 2 json schemas...
		modified, err = jsonpatch.MergePatch(stdout, patchJSON)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		re := regexp.MustCompile(`"required":(\[)["\p{L}\p{N},_-]+(\])`)
		bts, err = utils.PrettyString([]byte(re.ReplaceAllString(string(modified), `"required":$1$2`)))

		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		log.Info().Msgf("write schema into file: %s", schemaPath)
		err = ioutil.WriteFile(schemaPath, bts, 0644)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.schema-generator.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&overrideSchemaPath, "overSchemaPath", "o", defaultOverrideSchemaPath, color.BlueString("Path to json file with override schema values"))
	rootCmd.Flags().StringVarP(&valuesYamlPath, "valuesPath", "v", "", color.BlueString("Path to yaml/yml file with chart values"))
	rootCmd.Flags().StringVarP(&schemaPath, "schemaPath", "s", defaultSchemaDir, color.BlueString("Path for json schema file"))
}

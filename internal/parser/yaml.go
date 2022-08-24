package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/karuppiah7890/go-jsonschema-generator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/alex60217101990/json_schema_generator/internal/enums"
	"github.com/alex60217101990/json_schema_generator/internal/types"
	"github.com/alex60217101990/json_schema_generator/internal/utils"
)

// jessie ware - remember where you are

const (
	patternGroupName = "Json"

	defaultSchemaURL = "https://json-schema.org/draft/2019-09/schema"
)

type Parser struct {
	log    *zerolog.Logger
	regexp *regexp.Regexp

	patches         map[string]string
	patchesRequired map[string][]string

	errCh  chan error
	stopCh chan struct{}
}

func (p *Parser) ChangeAllPath(jsonSchema []byte) (data map[enums.PatchOperation]map[string]string) {
	data = make(map[enums.PatchOperation]map[string]string, 2)
	data[enums.Replace] = make(map[string]string, len(p.patches))
	data[enums.Add] = make(map[string]string, len(p.patchesRequired))

	var (
		correctPath string
		jsonPath    string
	)

	for path, val := range p.patchesRequired {
		correctPath = fmt.Sprintf("/properties/%s", strings.ReplaceAll(path, ".", "/"))
		jsonPath = fmt.Sprintf("properties.%s", path)

		m, ok := gjson.ParseBytes(jsonSchema).Get(jsonPath).Value().(map[string]interface{})
		if ok && len(m) > 0 {
			m["required"] = val

			bts, err := json.Marshal(m)
			if err != nil {
				p.log.Fatal().Stack().Err(errors.WithStack(err)).Msg("")
			}

			data[enums.Add][correctPath] = string(bts)
		}
	}

	for path, val := range p.patches {
		correctPath = fmt.Sprintf("/properties/%s", strings.ReplaceAll(path, ".", "/"))
		jsonPath = fmt.Sprintf("properties.%s", path)

		m, ok := gjson.ParseBytes(jsonSchema).Get(jsonPath).Value().(map[string]interface{})
		if ok && len(m) > 0 {
			v := make(map[string]interface{})
			err := json.Unmarshal([]byte(val), &v)
			if err != nil {
				p.log.Fatal().Stack().Err(errors.WithStack(err)).Msg("")
			}

			for key, value := range v {
				m[key] = value
			}

			newBts, err := json.Marshal(m)
			if err != nil {
				p.log.Fatal().Stack().Err(errors.WithStack(err)).Msg("")
			}

			data[enums.Replace][correctPath] = string(newBts)
			continue
		}

		data[enums.Replace][correctPath] = val
	}

	return data
}

func (p *Parser) ParseSync(node *yaml.Node, val []byte) (output []byte, err error) {
	out, errCh := p.ParseAsync(node, val)

	for {
		select {
		case output = <-out:
			return output, nil
		case err = <-errCh:
			return nil, err
		case <-p.stopCh:
			return output, nil
		}
	}
}

func (p *Parser) ParseAsync(node *yaml.Node, val []byte) (<-chan []byte, <-chan error) {
	outputCh := make(chan []byte)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				switch err := r.(type) {
				case string:
					p.errCh <- errors.WithStack(errors.New(err))
				case error:
					p.errCh <- errors.WithStack(err)
				}
			}

			close(outputCh)
			close(p.errCh)

			p.stopCh <- struct{}{}
			close(p.stopCh)
		}()

		p.start(node, "", false)

		var v chartutil.Values
		err := yaml.Unmarshal(val, &v)
		if err != nil {
			p.errCh <- errors.WithStack(err)
			return
		}

		schema := &jsonschema.Document{}
		schema.ReadDeep(&v)
		jsonBts, err := schema.Marshal()
		if err != nil {
			p.errCh <- errors.WithStack(err)
			return
		}

		var modified []byte

		patchesJSON := []*types.Patch{{
			OperationType: enums.Replace,
			Path:          "/$schema",
			Value:         []byte(fmt.Sprintf(`"%s"`, defaultSchemaURL)),
		}}

		patches := p.ChangeAllPath(jsonBts)

		for path, object := range patches[enums.Add] {
			patchesJSON = append(patchesJSON, &types.Patch{
				OperationType: enums.Add,
				Path:          path,
				Value:         []byte(object),
			})
		}

		for path, object := range patches[enums.Replace] {
			patchesJSON = append(patchesJSON, &types.Patch{
				OperationType: enums.Replace,
				Path:          path,
				Value:         []byte(object),
			})
		}

		tmpBts, err := json.Marshal(patchesJSON)
		if err != nil {
			p.errCh <- errors.WithStack(err)
			return
		}

		patch, err := jsonpatch.DecodePatch(tmpBts)
		if err != nil {
			p.errCh <- errors.WithStack(err)
			return
		}

		modified, err = patch.Apply(jsonBts)
		if err != nil {
			p.errCh <- errors.WithStack(err)
			return
		}

		bts, err := utils.PrettyString(modified)
		if err != nil {
			p.errCh <- errors.WithStack(err)
			return
		}

		outputCh <- bts
	}()

	return outputCh, p.errCh
}

func (p *Parser) Init(log *zerolog.Logger) *Parser {
	p.log = log
	p.patches = make(map[string]string)
	p.patchesRequired = make(map[string][]string)
	p.regexp = regexp.MustCompile(fmt.Sprintf(`@jsonSchema:\s?(?P<%s>{.*})`, patternGroupName))
	p.errCh = make(chan error, 10)
	p.stopCh = make(chan struct{})

	return p
}

func (p *Parser) start(n *yaml.Node, path string, isSlice bool) {
	newPath := path
	var checkSliceParent bool
	for i, node := range n.Content {
		if node.LineComment != "" && strings.Contains(node.LineComment, "@jsonSchema") {
			var (
				key    string
				value  string
				prefix string
			)

			if isSlice {
				prefix = "items.properties"
			} else {
				prefix = "properties"
			}

			if node.Value == "annotations" || n.Content[i-1].Value == "annotations" {
				p.log.Info().Msgf("comment: %s, path: %s", node.LineComment, key)
			}

			var fieldName string
			if node.Tag != "!!str" {
				key = fmt.Sprintf("%s.%s.%s", path, prefix, n.Content[i-1].Value)
				fieldName = n.Content[i-1].Value
			} else {
				key = fmt.Sprintf("%s.%s.%s", path, prefix, node.Value)
				fieldName = node.Value
			}

			val := make(map[string]interface{})
			match := p.regexp.FindStringSubmatch(node.LineComment)

			for i, name := range p.regexp.SubexpNames() {
				if name == patternGroupName && len(match[i]) > 0 {
					err := json.Unmarshal([]byte(match[i]), &val)
					if err != nil {
						p.errCh <- errors.WithStack(err)
					}

					value = match[i]
				}
			}

			p.log.Info().Msgf("comment: %s, path: %s", node.LineComment, key)
			p.patches[key] = value

			if v, ok := val["required"]; ok {
				err := p.appendRequired(v, true, path, fieldName)
				if err != nil {
					p.errCh <- errors.WithStack(err)
				}
			}

			if v, ok := val["optional"]; ok {
				err := p.appendRequired(v, false, path, fieldName)
				if err != nil {
					p.errCh <- errors.WithStack(err)
				}
			}
		}

		if node.Tag == "!!map" || node.Tag == "!!seq" {
			if isSlice {
				checkSliceParent = true
			}

			if (i > 0) && n.Content[i-1].Tag == "!!str" {
				if len(path) > 0 {
					newPath = fmt.Sprintf("%s.properties.%s", path, n.Content[i-1].Value)
				} else {
					newPath = n.Content[i-1].Value
				}
			}
		}

		if node.Tag == "!!seq" && len(path) > 0 && len(node.Content) > 0 {
			checkSliceParent = true
		}

		p.start(node, newPath, checkSliceParent)
	}
}

func (p *Parser) appendRequired(value interface{}, requiredField bool, path, fieldName string) (err error) {
	switch requiredInfo := value.(type) {
	case string:
		var boolValue bool
		boolValue, err = strconv.ParseBool(requiredInfo)
		if err != nil {
			return err
		}

		if requiredField {
			if boolValue && len(path) > 0 {
				p.patchesRequired[path] = append(p.patchesRequired[path], fieldName)
			}

			return
		}

		if !boolValue && len(path) > 0 {
			p.patchesRequired[path] = append(p.patchesRequired[path], fieldName)
		}
	case bool:
		if requiredField {
			if requiredInfo && len(path) > 0 {
				p.patchesRequired[path] = append(p.patchesRequired[path], fieldName)
			}

			return
		}

		if !requiredInfo && len(path) > 0 {
			p.patchesRequired[path] = append(p.patchesRequired[path], fieldName)
		}
	}

	return
}

package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

const (
	patternGroupName = "Json"
)

type Parser struct {
	log     *zerolog.Logger
	regexp  *regexp.Regexp
	patches map[string]string
}

func (p *Parser) PatchesList() map[string]string {
	return p.patches
}

func (p *Parser) Length() int {
	return len(p.patches)
}

func (p *Parser) ChangeAllPath(jsonSchema []byte) (data map[string]string) {
	data = make(map[string]string, len(p.patches))
	for path, val := range p.patches {
		var (
			correctPath string
			jsonPath    string
		)

		//for _, field := range strings.Split(path, ".properties") {
		//	correctPath = fmt.Sprintf("%s/properties/%s", correctPath, field)
		//	jsonPath = fmt.Sprintf("%s.properties.%s", jsonPath, field)
		//}

		////correctPath = strings.TrimPrefix(correctPath, "/")
		//jsonPath = strings.TrimPrefix(jsonPath, ".")

		correctPath = fmt.Sprintf("/properties/%s", strings.ReplaceAll(path, ".", "/"))
		jsonPath = fmt.Sprintf("properties.%s", path)

		m, ok := gjson.ParseBytes(jsonSchema).Get(jsonPath).Value().(map[string]interface{})
		if ok && len(m) > 0 {
			v := make(map[string]interface{})
			err := json.Unmarshal([]byte(val), &v)
			if err != nil {
				p.log.Fatal().Msg(err.Error())
			}

			for key, value := range v {
				m[key] = value
			}

			newBts, err := json.Marshal(m)
			if err != nil {
				p.log.Fatal().Msg(err.Error())
			}

			data[correctPath] = string(newBts)
			continue
		}

		data[correctPath] = val
	}

	return data
}

func (p *Parser) Init(log *zerolog.Logger) *Parser {
	p.log = log
	p.patches = make(map[string]string)
	p.regexp = regexp.MustCompile(fmt.Sprintf(`@jsonSchema:\s?(?P<%s>{.*})`, patternGroupName))
	return p
}

func (p *Parser) Start(n *yaml.Node, path string, isSlice bool) {
	newPath := path
	var checkSliceParent bool
	for i, node := range n.Content {
		if node.LineComment != "" && strings.Contains(node.LineComment, "@jsonSchema") {
			var (
				key     string
				value   string
				preffix string
			)

			if isSlice {
				preffix = "items.properties"
			} else {
				preffix = "properties"
			}

			if node.Value == "annotations" || n.Content[i-1].Value == "annotations" {
				p.log.Info().Msgf("comment: %s, path: %s", node.LineComment, key)
			}

			if node.Tag != "!!str" {
				key = fmt.Sprintf("%s.%s.%s", path, preffix, n.Content[i-1].Value)
			} else {
				key = fmt.Sprintf("%s.%s.%s", path, preffix, node.Value)
			}

			match := p.regexp.FindStringSubmatch(node.LineComment)
			for i, name := range p.regexp.SubexpNames() {
				if name == patternGroupName && len(match[i]) > 0 {
					val := make(map[string]interface{})
					err := json.Unmarshal([]byte(match[i]), &val)
					if err != nil {
						p.log.Fatal().Msg(err.Error())
					}

					value = match[i]
					//value["type"] = ""
				}
			}

			p.log.Info().Msgf("comment: %s, path: %s", node.LineComment, key)
			p.patches[key] = value
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

		p.Start(node, newPath, checkSliceParent)
	}
}

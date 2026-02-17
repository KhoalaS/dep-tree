//nolint:govet
package js_grammar

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/gabotechs/dep-tree/internal/language"

	"github.com/gabotechs/dep-tree/internal/utils"
)

var vueScriptSetupRegex = regexp.MustCompile(`(?s)<script setup.+?>\n(.+)</script>`)

type Statement struct {
	// imports.
	DynamicImport *DynamicImport `  @@`
	StaticImport  *StaticImport  `| @@`
	Require       *Require       `| @@`
	// exports.
	DeclarationExport *DeclarationExport `| @@`
	DefaultExport     *DefaultExport     `| @@`
	ProxyExport       *ProxyExport       `| @@`
	ListExport        *ListExport        `| @@`
}

type File struct {
	Statements []*Statement `(@@ | ANY | ALL | Punct | Ident | String | BacktickString)*`
}

var (
	lex = lexer.MustSimple(
		[]lexer.SimpleRule{
			{"ALL", `\*`},
			{"Punct", `[:,{}()]`},
			{"Ident", `[_$a-zA-Z][_$a-zA-Z0-9]*`},
			{"String", `'(?:\\.|[^'])*'` + "|" + `"(?:\\.|[^"])*"`},
			{"BacktickString", "`(?:\\\\.|[^`])*`"},
			{"Comment", `//.*|/\*(.|\n)*?\*/`},
			{"Whitespace", `\s+`},
			{"ANY", `.`},
		},
	)
	parser = participle.MustBuild[File](
		participle.Lexer(lex),
		participle.Elide("Whitespace", "Comment"),
		utils.UnquoteSafe("String"),
		participle.UseLookahead(1024),
	)
)

func Parse(filePath string) (*language.FileInfo, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(filePath)
	if ext == ".vue" {
		vueScriptMatches := vueScriptSetupRegex.FindSubmatch(content)
		if len(vueScriptMatches) != 2 {
			return nil, fmt.Errorf("vue file does not use script setup syntax")
		}

		content = vueScriptMatches[1]
	}

	statements, err := parser.ParseBytes(filePath, content)
	if err != nil {
		return nil, err
	}
	return &language.FileInfo{
		Content: statements,
		Loc:     bytes.Count(content, []byte("\n")),
		Size:    len(content),
		AbsPath: filePath,
	}, nil
}

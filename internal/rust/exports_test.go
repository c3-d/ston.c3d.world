package rust

import (
	"context"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gabotechs/dep-tree/internal/language"
)

func TestLanguage_ParseExports(t *testing.T) {
	absTestFolder, _ := filepath.Abs(path.Join(testFolder))

	tests := []struct {
		Name     string
		Expected []language.ExportEntry
		Errors   []error
	}{
		{
			Name: "lib.rs",
			Expected: []language.ExportEntry{
				{
					Names: []language.ExportName{{Original: "div"}},
					Path:  path.Join(absTestFolder, "src", "lib.rs"),
				},
				{
					Names: []language.ExportName{{Original: "abs"}},
					Path:  path.Join(absTestFolder, "src", "abs", "abs.rs"),
				},
				{
					Names: []language.ExportName{{Original: "div"}},
					Path:  path.Join(absTestFolder, "src", "div", "mod.rs"),
				},
				{
					Names: []language.ExportName{{Original: "avg"}},
					Path:  path.Join(absTestFolder, "src", "avg_2.rs"),
				},
				{
					Names: []language.ExportName{{Original: "sum"}},
					Path:  path.Join(absTestFolder, "src", "lib.rs"),
				},
				{
					All:  true,
					Path: path.Join(absTestFolder, "src", "sum.rs"),
				},
				{
					Names: []language.ExportName{{Original: "run"}},
					Path:  path.Join(absTestFolder, "src", "lib.rs"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			a := require.New(t)
			_, _lang, err := MakeRustLanguage(context.Background(), path.Join(testFolder, "src", "lib.rs"), nil)
			a.NoError(err)

			lang := _lang.(*Language)

			file, err := lang.ParseFile(path.Join(absTestFolder, "src", tt.Name))
			a.NoError(err)

			exports, err := lang.ParseExports(file)
			a.NoError(err)
			a.Equal(tt.Expected, exports.Exports)
			a.Equal(tt.Errors, exports.Errors)
		})
	}
}

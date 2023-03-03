package js

import (
	"context"
	"path"

	"dep-tree/internal/js/js_grammar"
	"dep-tree/internal/language"
)

type ExportsCacheKey string

func (l *Language) ParseExports(ctx context.Context, file *js_grammar.File) (context.Context, *language.ExportsResult, error) {
	exports := make([]language.ExportEntry, 0)
	var errors []error

	for _, stmt := range file.Statements {
		switch {
		case stmt == nil:
			// Is this even possible?
		case stmt.DeclarationExport != nil:
			exports = append(exports, language.ExportEntry{
				Names: []language.ExportName{
					{
						Original: stmt.DeclarationExport.Name,
					},
				},
				Id: file.Path,
			})
		case stmt.ListExport != nil:
			if stmt.ListExport.ExportDeconstruction != nil {
				for _, name := range stmt.ListExport.ExportDeconstruction.Names {
					exports = append(exports, language.ExportEntry{
						Names: []language.ExportName{
							{
								Original: name.Original,
								Alias:    name.Alias,
							},
						},
						Id: file.Path,
					})
				}
			}
		case stmt.DefaultExport != nil:
			if stmt.DefaultExport.Default {
				exports = append(exports, language.ExportEntry{
					Names: []language.ExportName{
						{
							Original: "default",
						},
					},
					Id: file.Path,
				})
			}
		case stmt.ProxyExport != nil:
			exportFrom, err := l.ResolvePath(stmt.ProxyExport.From, path.Dir(file.Path))
			if err != nil {
				errors = append(errors, err)
				continue
			}

			switch {
			case stmt.ProxyExport.ExportAll:
				if stmt.ProxyExport.ExportAllAlias != "" {
					exports = append(exports, language.ExportEntry{
						Names: []language.ExportName{
							{
								Original: stmt.ProxyExport.ExportAllAlias,
							},
						},
						Id: exportFrom,
					})
				} else {
					exports = append(exports, language.ExportEntry{
						All: true,
						Id:  exportFrom,
					})
				}
			case stmt.ProxyExport.ExportDeconstruction != nil:
				names := make([]language.ExportName, 0)
				for _, name := range stmt.ProxyExport.ExportDeconstruction.Names {
					names = append(names, language.ExportName{
						Original: name.Original,
						Alias:    name.Alias,
					})
				}

				exports = append(exports, language.ExportEntry{
					Names: names,
					Id:    exportFrom,
				})
			}
		}
	}
	return ctx, &language.ExportsResult{
		Exports: exports,
		Errors:  errors,
	}, nil
}

// Package migrations expõe os arquivos SQL embutidos para uso por goose
// (em dev e nos testes), independente do diretório de trabalho.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

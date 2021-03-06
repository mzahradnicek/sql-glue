package sqlg

import (
	"os"
)

type Builder struct {
	cfg Config
}

func (b *Builder) Glue(q *Qg) (string, []interface{}, error) {
	return q.ToSql(&b.cfg)
}

func NewBuilder(cfg Config) *Builder {
	res := &Builder{}

	if cfg.IdentifierEscape == nil {
		os.Stderr.WriteString("Error: sqlg config IdentifierEscape cant be nil\n")
		os.Exit(1)
	}

	if cfg.KeyModifier == nil {
		os.Stderr.WriteString("Error: sqlg config KeyModifier cant be nil\n")
		os.Exit(1)
	}

	if cfg.Tag == "" {
		cfg.Tag = "sqlg"
	}

	res.cfg = cfg

	return res
}

package sqlg

import (
	"os"
)

type Builder struct {
	cfg      *Config
	splitter *Splitter
}

func (b *Builder) Glue(q *Qg) (string, []interface{}, error) {
	return q.Process(b.cfg)
}

func (b *Builder) Split(i interface{}, exclude ...string) ([]string, []interface{}, error) {
	return b.splitter.Split(i, exclude...)
}

func NewBuilder(cfg *Config) *Builder {
	res := &Builder{}

	if cfg.IdentifierEscape == nil {
		os.Stderr.WriteString("Error: sql-glue config IdentifierEscape cant be nil\n")
		os.Exit(1)
	}

	if cfg.KeyModifier == nil {
		os.Stderr.WriteString("Error: sql-glue config KeyModifier cant be nil\n")
		os.Exit(1)
	}

	res.splitter = NewSplitter().KeyModifier(cfg.KeyModifier).Tag(cfg.Tag)

	if cfg.Tag == "" {
		cfg.Tag = "sqlg"
	}

	res.cfg = cfg

	return res
}

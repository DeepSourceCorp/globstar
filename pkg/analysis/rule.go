package analysis

import sitter "github.com/smacker/go-tree-sitter"

type VisitFn func(checker Checker, analyzer *Analyzer, node *sitter.Node)

type Checker interface {
	NodeType() string
	GetLanguage() Language
	OnEnter() *VisitFn
	OnLeave() *VisitFn
}

type checkerImpl struct {
	nodeType string
	language Language
	onEnter  *VisitFn
	onLeave  *VisitFn
}

func (r *checkerImpl) NodeType() string      { return r.nodeType }
func (r *checkerImpl) GetLanguage() Language { return r.language }
func (r *checkerImpl) OnEnter() *VisitFn     { return r.onEnter }
func (r *checkerImpl) OnLeave() *VisitFn     { return r.onLeave }

func CreateChecker(nodeType string, language Language, onEnter, onLeave *VisitFn) Checker {
	return &checkerImpl{
		nodeType: nodeType,
		language: language,
		onEnter:  onEnter,
		onLeave:  onLeave,
	}
}

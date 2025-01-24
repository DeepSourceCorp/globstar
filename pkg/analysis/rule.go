package analysis

import sitter "github.com/smacker/go-tree-sitter"

type VisitFn func(rule Rule, analyzer *Analyzer, node *sitter.Node)

type Rule interface {
	NodeType() string
	GetLanguage() Language
	OnEnter() *VisitFn
	OnLeave() *VisitFn
}

type ruleImpl struct {
	nodeType string
	language Language
	onEnter  *VisitFn
	onLeave  *VisitFn
}

func (r *ruleImpl) NodeType() string      { return r.nodeType }
func (r *ruleImpl) GetLanguage() Language { return r.language }
func (r *ruleImpl) OnEnter() *VisitFn     { return r.onEnter }
func (r *ruleImpl) OnLeave() *VisitFn     { return r.onLeave }

func CreateRule(nodeType string, language Language, onEnter, onLeave *VisitFn) Rule {
	return &ruleImpl{
		nodeType: nodeType,
		language: language,
		onEnter:  onEnter,
		onLeave:  onLeave,
	}
}

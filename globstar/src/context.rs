use std::{cell::RefCell, fmt, rc::Rc};

use scope_resolution::{LocalScope, ScopeStack};
use tree_sitter::{Node, Range, Tree};

/// Pass around scope and injection data across lint rules.
#[derive(Debug)]
pub struct Context<'a> {
    pub(crate) root_scope: Rc<RefCell<LocalScope<'a>>>,
    pub(crate) injected_trees: Vec<InjectedTree>,
}

pub struct InjectedTree {
    pub tree: Tree,
    pub original_range: Range,
}

impl fmt::Debug for InjectedTree {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("InjectedTree")
            .field("tree", &self.tree.root_node().range())
            .field("range", &self.original_range)
            .finish()
    }
}

impl<'a> Context<'a> {
    /// Produce the top-most scope for a given file
    pub fn root_scope(&self) -> Rc<RefCell<LocalScope<'a>>> {
        Rc::clone(&self.root_scope)
    }

    /// Produce the the nearest scope that can fully contain `range`
    pub fn scope_by_range(&self, range: &Range) -> Option<Rc<RefCell<LocalScope<'a>>>> {
        scope_resolution::scope_by_range(self.root_scope(), range)
    }

    /// Produce the scope that `node` belongs to
    pub fn scope_of(&self, node: Node) -> Option<Rc<RefCell<LocalScope<'a>>>> {
        self.scope_by_range(&node.range())
    }

    /// Produce a list of scopes upto the root scope, that can fully contain `range`
    pub fn scope_stack_by_range(&self, range: &Range) -> Option<ScopeStack<'a>> {
        Some(scope_resolution::scope_stack(self.scope_by_range(range)?))
    }

    /// Produce a list of scopes upto the root scope, that `node` belongs to
    pub fn scope_stack_of(&self, node: Node) -> Option<ScopeStack<'a>> {
        self.scope_stack_by_range(&node.range())
    }

    /// Produce the first injected syntax tree that is fully contained in `range`
    pub fn injected_tree_by_range(&self, range: &Range) -> Option<&InjectedTree> {
        self.injected_trees
            .iter()
            .find(|t| contains_range(&t.original_range, range))
    }

    /// Produce the first injected syntax tree that is fully contained by `node`
    pub fn injected_tree_of(&self, node: &Node) -> Option<&InjectedTree> {
        self.injected_tree_by_range(&node.range())
    }
}

fn contains_range(a: &Range, b: &Range) -> bool {
    b.start_byte >= a.start_byte && b.end_byte <= a.end_byte
}

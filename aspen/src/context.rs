use std::{cell::RefCell, rc::Rc};

use cedar::{LocalScope, ScopeStack};
use tree_sitter::{Node, Range, Tree};

pub struct Context<'a> {
    pub(crate) root_scope: Rc<RefCell<LocalScope<'a>>>,
    pub(crate) injected_trees: Vec<InjectedTree>,
}

pub struct InjectedTree {
    pub tree: Tree,
    pub original_range: Range,
}

impl<'a> Context<'a> {
    pub fn root_scope(&self) -> Rc<RefCell<LocalScope<'a>>> {
        Rc::clone(&self.root_scope)
    }
    pub fn scope_by_range(&self, range: &Range) -> Option<Rc<RefCell<LocalScope<'a>>>> {
        cedar::scope_by_range(self.root_scope(), range)
    }
    pub fn scope_of(&self, node: Node) -> Option<Rc<RefCell<LocalScope<'a>>>> {
        self.scope_by_range(&node.range())
    }
    pub fn scope_stack_by_range(&self, range: &Range) -> Option<ScopeStack<'a>> {
        Some(cedar::scope_stack(self.scope_by_range(range)?))
    }
    pub fn scope_stack_of(&self, node: Node) -> Option<ScopeStack<'a>> {
        self.scope_stack_by_range(&node.range())
    }
    pub fn injected_tree_by_range(&self, range: &Range) -> Option<&InjectedTree> {
        self.injected_trees
            .iter()
            .find(|t| contains_range(&t.original_range, range))
    }
    pub fn injected_tree_of(&self, node: &Node) -> Option<&InjectedTree> {
        self.injected_tree_by_range(&node.range())
    }
}

fn contains_range(a: &Range, b: &Range) -> bool {
    b.start_byte >= a.start_byte && b.end_byte <= a.end_byte
}

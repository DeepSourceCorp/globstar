use crate::LocalDef;

use std::{cell::RefCell, fmt, rc::Rc};

use tree_sitter::Range;

pub struct LocalScope<'a> {
    pub range: Range,
    pub local_defs: Vec<Rc<RefCell<LocalDef<'a>>>>,
    pub child_scopes: Vec<Rc<RefCell<LocalScope<'a>>>>,
    pub parent_scope: Option<Rc<RefCell<LocalScope<'a>>>>,
}

impl<'a> LocalScope<'a> {
    pub fn new(range: &Range) -> Self {
        Self {
            range: range.clone(),
            local_defs: Vec::new(),
            child_scopes: Vec::new(),
            parent_scope: None,
        }
    }
}

pub fn scope_stack<'a>(start_scope: Rc<RefCell<LocalScope<'a>>>) -> ScopeStack<'a> {
    ScopeStack {
        start_scope: Some(start_scope),
    }
}

pub struct ScopeStack<'a> {
    start_scope: Option<Rc<RefCell<LocalScope<'a>>>>,
}

impl<'a> Iterator for ScopeStack<'a> {
    type Item = Rc<RefCell<LocalScope<'a>>>;
    fn next(&mut self) -> Option<Self::Item> {
        if let Some(start_scope) = &self.start_scope {
            let parent = start_scope
                .borrow()
                .parent_scope
                .as_ref()
                .map(|t| Rc::clone(t));
            let original = Rc::clone(start_scope);
            self.start_scope = parent;
            return Some(original);
        } else {
            None
        }
    }
}

// curtom Debug impl because printing parent_scope causes infinite recursion
impl<'a> fmt::Debug for LocalScope<'a> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("LocalScope")
            .field("range", &self.range)
            .field("local_defs", &self.local_defs)
            .field("child_scopes", &self.child_scopes)
            .field("parent_scope", &self.parent_scope.is_some())
            .finish()
    }
}

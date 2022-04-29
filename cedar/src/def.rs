use crate::LocalScope;

use std::{cell::RefCell, fmt, ops, rc::Rc};

use tree_sitter::Range;

pub struct LocalDef<'a> {
    pub name: &'a str,
    pub is_mutable: bool,
    pub def_range: Range,
    pub value_range: Option<Range>,
    pub scope: Rc<RefCell<LocalScope<'a>>>,
    pub references: Vec<Rc<Reference<'a>>>,
}

impl<'a> fmt::Debug for LocalDef<'a> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("LocalDef")
            .field("def_range", &self.def_range)
            .field("value_range", &self.value_range)
            .field("is_mutable", &self.is_mutable)
            .field("name", &self.name)
            .field("owning scope", &self.scope.borrow().range)
            .field("reference count", &self.references.len())
            .field(
                "references",
                &self
                    .references
                    .iter()
                    .map(|reference| reference.range.clone())
                    .collect::<Vec<_>>(),
            )
            .finish()
    }
}

impl<'a> LocalDef<'a> {
    pub fn new(name: &'a str, def_range: &Range, scope: Rc<RefCell<LocalScope<'a>>>) -> Self {
        Self {
            name,
            is_mutable: false,
            def_range: def_range.clone(),
            value_range: None,
            references: Vec::new(),
            scope,
        }
    }
    pub fn new_with_value(
        name: &'a str,
        def_range: &Range,
        value_range: &Range,
        scope: Rc<RefCell<LocalScope<'a>>>,
    ) -> Self {
        Self {
            name,
            is_mutable: false,
            def_range: def_range.clone(),
            value_range: Some(value_range.clone()),
            references: Vec::new(),
            scope,
        }
    }
}

pub struct Reference<'a> {
    pub range: Range,
    pub original_def: Rc<RefCell<LocalDef<'a>>>,
}

impl<'a> fmt::Debug for Reference<'a> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("Reference")
            .field("range", &self.range)
            .field("original_def", &self.original_def.borrow().def_range)
            .finish()
    }
}

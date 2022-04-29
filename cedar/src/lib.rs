mod def;
mod scope;
mod utils;

pub use def::{LocalDef, Reference};
pub use scope::{scope_stack, LocalScope, ScopeStack};
pub use utils::scope_by_range;

use std::{cell::RefCell, rc::Rc};

use tree_sitter::{Node, Query, QueryCursor};

#[non_exhaustive]
#[derive(Debug, PartialEq, Eq, Clone, Copy)]
pub enum ResolutionMethod {
    Generic,
}

impl ResolutionMethod {
    pub fn build_scope<'a>(
        &self,
        query: &Query,
        root_node: Node,
        src: &'a str,
    ) -> Rc<RefCell<LocalScope<'a>>> {
        match self {
            ResolutionMethod::Generic => scope_res_generic(query, root_node, src),
        }
    }
}

fn scope_res_generic<'a>(
    query: &Query,
    root_node: Node,
    src: &'a str,
) -> Rc<RefCell<LocalScope<'a>>> {
    let mut local_def_capture_index = None;
    let mut local_def_value_capture_index = None;
    let mut local_ref_capture_index = None;
    let mut local_scope_capture_index = None;
    for (i, name) in query.capture_names().iter().enumerate() {
        let i = Some(i as u32);
        match name.as_str() {
            "local.definition" => local_def_capture_index = i,
            "local.definition-value" => local_def_value_capture_index = i,
            "local.reference" => local_ref_capture_index = i,
            "local.scope" => local_scope_capture_index = i,
            _ => {}
        }
    }
    let mut cursor = QueryCursor::new();
    let captures = cursor.captures(&query, root_node, src.as_bytes());

    let root_scope = Rc::new(RefCell::new(LocalScope {
        range: root_node.range(),
        local_defs: Vec::new(),
        child_scopes: Vec::new(),
        parent_scope: None,
    }));

    for (match_, capture_index) in captures {
        let capture = match_.captures[capture_index];
        let range = capture.node.range();
        let name = &src[range.start_byte..range.end_byte];

        if local_scope_capture_index == Some(capture.index) {
            utils::insert_scope(Rc::clone(&root_scope), &range);
        } else if local_def_capture_index == Some(capture.index) {
            utils::insert_def(Rc::clone(&root_scope), name, &range, None);
        } else if local_ref_capture_index == Some(capture.index) {
            utils::insert_ref(Rc::clone(&root_scope), name, &range);
        }
    }
    root_scope
}

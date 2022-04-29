use crate::{scope::scope_stack, LocalDef, LocalScope, Reference};

use std::{cell::RefCell, rc::Rc};

use tree_sitter::Range;

fn contains_range(a: &Range, b: &Range) -> bool {
    b.start_byte >= a.start_byte && b.end_byte <= a.end_byte
}

pub fn insert_scope<'a>(target: Rc<RefCell<LocalScope<'a>>>, range: &Range) -> bool {
    let scope = Rc::new(RefCell::new(LocalScope::new(range)));
    let target_range = target.borrow().range.clone();
    let scope_range = scope.borrow().range.clone();
    if contains_range(&target_range, &scope_range) {
        for child_scope in target.borrow_mut().child_scopes.iter() {
            if insert_scope(Rc::clone(child_scope), range) {
                return true;
            }
        }
        target.borrow_mut().child_scopes.push(Rc::clone(&scope));
        scope.borrow_mut().parent_scope = Some(Rc::clone(&target));
        return true;
    }
    false
}

pub fn scope_by_range<'a>(
    target: Rc<RefCell<LocalScope<'a>>>,
    range: &Range,
) -> Option<Rc<RefCell<LocalScope<'a>>>> {
    let target_range = target.borrow().range.clone();
    if contains_range(&target_range, range) {
        for child_scope in target.borrow_mut().child_scopes.iter() {
            if let Some(t) = scope_by_range(Rc::clone(&child_scope), range) {
                return Some(t);
            }
        }
        return Some(Rc::clone(&target));
    }
    None
}

pub fn insert_def<'a>(
    target: Rc<RefCell<LocalScope<'a>>>,
    name: &'a str,
    def_range: &Range,
    value_range: Option<&Range>,
) -> bool {
    let target_range = target.borrow().range.clone();
    if contains_range(&target_range, def_range) {
        for child_scope in target.borrow_mut().child_scopes.iter() {
            if insert_def(Rc::clone(child_scope), name, def_range, value_range) {
                return true;
            }
        }
        let def = value_range
            .map(|vr| {
                Rc::new(RefCell::new(LocalDef::new_with_value(
                    name,
                    def_range,
                    vr,
                    Rc::clone(&target),
                )))
            })
            .unwrap_or_else(|| {
                Rc::new(RefCell::new(LocalDef::new(
                    name,
                    def_range,
                    Rc::clone(&target),
                )))
            });
        target.borrow_mut().local_defs.push(def);
        return true;
    }
    false
}

pub fn insert_ref<'a>(root_scope: Rc<RefCell<LocalScope<'a>>>, name: &str, range: &Range) -> bool {
    if let Some(local_scope) = scope_by_range(root_scope, &range) {
        // find the original def in the reverse ordor of scopes up to the root scope
        for scope in scope_stack(local_scope) {
            for local_def in scope.borrow().local_defs.iter() {
                if local_def.borrow().name == name {
                    let reference = Rc::new(Reference {
                        range: range.clone(),
                        original_def: Rc::clone(&local_def),
                    });
                    local_def.borrow_mut().references.push(reference);
                    return true;
                }
            }
        }
    }
    false
}

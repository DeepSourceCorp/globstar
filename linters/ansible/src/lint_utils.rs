use aspen::tree_sitter::QueryCapture;

pub fn is_ansible_builtin(ident: &str, builtin: &'static str) -> bool {
    match ident.strip_prefix("ansible.builtin") {
        Some(rest) => rest == builtin,
        None => ident == builtin,
    }
}

pub trait AsText {
    fn as_text<'a>(&self, src: &'a [u8]) -> Option<&'a str>;
}

impl<'q> AsText for QueryCapture<'q> {
    fn as_text<'a>(&self, src: &'a [u8]) -> Option<&'a str> {
        self.node.utf8_text(src).ok()
    }
}

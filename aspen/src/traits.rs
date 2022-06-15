use tree_sitter::{QueryCapture, QueryMatches, TextProvider};

pub trait MapCapture {
    fn map_capture<B>(self, capture_name: &str, f: impl Fn(&QueryCapture) -> B) -> Vec<B>;
    fn flat_map_capture<B>(
        self,
        capture_name: &str,
        f: impl Fn(&QueryCapture) -> Option<B>,
    ) -> Vec<B>;
}

impl<'a, 'tree, T> MapCapture for QueryMatches<'a, 'tree, T>
where
    T: TextProvider<'a>,
{
    fn map_capture<B>(self, capture_name: &str, f: impl Fn(&QueryCapture) -> B) -> Vec<B> {
        let capture_idx = self
            .query()
            .capture_index_for_name(capture_name)
            .expect("u r idiot");
        self.flat_map(|m| m.captures.iter().filter(|c| c.index == capture_idx))
            .map(f)
            .collect()
    }

    fn flat_map_capture<B>(
        self,
        capture_name: &str,
        f: impl Fn(&QueryCapture) -> Option<B>,
    ) -> Vec<B> {
        let capture_idx = self
            .query()
            .capture_index_for_name(capture_name)
            .expect("u r idiot");
        self.flat_map(|m| m.captures.iter().filter(|c| c.index == capture_idx))
            .flat_map(f)
            .collect()
    }
}

use tree_sitter::{Node, Query, QueryCapture, QueryCursor, QueryMatches, TextProvider};

/// Utility trait to iterate over captures.
///
/// This trait never needs to be implemented upon a type, consider using
/// the provided implementation for `QueryMatches`.
pub trait MapCapture {
    /// Apply `f` to every capture by name, `capture_name`.
    fn map_capture<B>(self, capture_name: &str, f: impl Fn(&QueryCapture) -> B) -> Vec<B>;

    /// Works like `map_capture`, but returns the `value`s for which
    /// the closure `f` returns `Some(v)`.
    fn filter_map_capture<B>(
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

    fn filter_map_capture<B>(
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

/// Utility trait to verify if a given query had any matches over
/// some source text.
///
/// This trait never needs to be implemented upon a type,
/// consider using the provided implementation for `QueryCursor`.
pub trait IsMatch {
    /// This function is a shorthand for
    /// `QueryCursor(..).matches(query, node, src).next().is_some`.
    fn is_match<'a, 'tree: 'a, T>(
        &'a mut self,
        query: &'a Query,
        node: Node<'tree>,
        src: T,
    ) -> bool
    where
        T: TextProvider<'a> + 'a;
}

impl IsMatch for QueryCursor {
    fn is_match<'a, 'tree: 'a, T>(&'a mut self, query: &'a Query, node: Node<'tree>, src: T) -> bool
    where
        T: TextProvider<'a> + 'a,
    {
        self.matches(query, node, src).next().is_some()
    }
}

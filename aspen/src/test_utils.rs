use crate::Linter;

impl Linter {
    /// Test the linter on an annotated source file
    pub fn test(&self, src: &str) {
        use crate::Diagnostic;

        use holly::{extract_annotations, trim_indent};
        use pretty_assertions::assert_eq as pretty_assert_eq;

        let src = trim_indent(src);
        let annotations = extract_annotations(&src, self.comment_str);
        let diagnostics = self.__analyze(&src);

        // panic if there are more annotations than diagnostics produced
        // or vice-versa
        let al = annotations.len();
        let dl = diagnostics.len();
        assert_eq!(al, dl, "Annotations ({}), Diagnostics ({})", al, dl);

        for (annotation, diagnostic) in annotations.iter().zip(diagnostics) {
            let ((range, content), comment) = annotation;
            let (_, Diagnostic { at, message }) = diagnostic;

            let actual_content = &src[at.start_byte..at.end_byte];

            pretty_assert_eq!(
                comment.as_str(),
                message.as_str(),
                "Comment `{}`, Message `{}`",
                comment,
                message
            );
            pretty_assert_eq!(actual_content, content);
            assert_eq!(*range, (at.start_byte, at.end_byte),);
        }
    }
}

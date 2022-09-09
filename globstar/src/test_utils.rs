use crate::Linter;

impl Linter {
    /// Test the linter on an annotated source file
    pub fn test(&self, src: &str) {
        use crate::{Diagnostic, Occurrence};

        use pretty_assertions::assert_eq as pretty_assert_eq;
        use test_utils::{extract_annotations, trim_indent};

        let src = trim_indent(src);
        let annotations = extract_annotations(&src, self.comment_str);
        let occurrences = self.__analyze(&src);

        // panic if there are more annotations than diagnostics produced
        // or vice-versa
        let al = annotations.len();
        let ol = occurrences.len();
        assert_eq!(
            al,
            ol,
            "Annotations ({}), Occurrences ({}):\n{}",
            al,
            ol,
            occurrences
                .iter()
                .map(ToString::to_string)
                .collect::<Vec<_>>()
                .join("\n")
        );

        for (annotation, occurrence) in annotations.iter().zip(occurrences) {
            let ((range, content), comment) = annotation;
            let Occurrence {
                diagnostic: Diagnostic { at, message },
                ..
            } = occurrence;

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

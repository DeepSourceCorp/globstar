pub fn trim_indent(mut text: &str) -> String {
    if text.starts_with('\n') {
        text = &text[1..];
    }
    let indent = text
        .lines()
        .filter(|it| !it.trim().is_empty())
        .map(|it| it.len() - it.trim_start().len())
        .min()
        .unwrap_or(0);
    text.split_inclusive('\n')
        .map(|line| {
            if line.len() <= indent {
                line.trim_start_matches(' ')
            } else {
                &line[indent..]
            }
        })
        .collect()
}

type Range = (usize, usize);

pub fn extract_range_annotation(mut line: &str) -> Option<Range> {
    let marker = '^';

    line.find(marker).map(|idx| {
        line = &line[idx..];
        let len = line.chars().take_while(|&it| it == marker).count();

        (idx, idx + len)
    })
}

type Annotation = (Range, String);

pub fn extract_annotations(src: &str, comment_start: &str) -> Vec<(Annotation, String)> {
    let mut res = Vec::new();
    let mut prev_line_start: Option<usize> = None;
    let mut line_start = 0usize;
    for line in src.split_inclusive('\n') {
        if let Some(idx) = line.find(comment_start) {
            let annotation_offset = line[..idx + comment_start.len()].len();
            if let Some(mut range) = extract_range_annotation(&line[idx + comment_start.len()..]) {
                let comment = line[idx + comment_start.len() + range.1..]
                    .trim()
                    .to_owned();
                range = (
                    range.0 + prev_line_start.unwrap() + annotation_offset,
                    range.1 + prev_line_start.unwrap() + annotation_offset,
                );
                let annotation = (range, String::from(&src[range.0..range.1]));
                res.push((annotation, comment));
            }
        }
        prev_line_start = Some(line_start);
        line_start += line.len();
    }
    res
}

// test the test suite
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_range_extraction() {
        let line = "  ^^^^^^^^^";
        let range = extract_range_annotation(&line);
        assert!(range.is_some());
        assert_eq!(range.unwrap(), (2, 11))
    }

    #[test]
    fn test_annotation_extractions() {
        let src = trim_indent(
            r#"
            abcde
          # ^^^^^ some text here
        "#,
        );
        let range = extract_annotations(&src, "#");
        assert!(!range.is_empty());
        let (annotation, comment) = range.get(0).unwrap();
        let (range, content) = annotation;
        assert_eq!(*range, (2, 7));
        assert_eq!(content, "abcde");
        assert_eq!(comment, "some text here");
    }

    #[test]
    fn test_annotation_extractions_non_trivial() {
        let src = dbg!(trim_indent(
            r#"
            if abc = 2
             # ^^^^^^^ assignment instead of eq
                puts "hello"
            else
                puts
              # ^^^^ empty `puts`
        "#,
        ));
        let annotations = extract_annotations(&src, "#");

        assert!(!annotations.is_empty());

        let (annotation, comment) = annotations.get(0).unwrap();
        let (range, content) = annotation;
        assert_eq!(*range, (3, 10));
        assert_eq!(content, "abc = 2");
        assert_eq!(comment, "assignment instead of eq");

        let (annotation, comment) = annotations.get(1).unwrap();
        let (range, content) = annotation;
        assert_eq!(*range, (73, 77));
        assert_eq!(content, "puts");
        assert_eq!(comment, "empty `puts`");
    }
}

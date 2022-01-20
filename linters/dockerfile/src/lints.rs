macro_rules! lints {
    ($($s:ident),*,) => {
        lints!($($s),*);
    };
    ($($s:ident),*) => {
        $(
            mod $s;
        )*
        ::lazy_static::lazy_static! {
            pub static ref LINTS: Vec<&'static aspen::Lint> = vec![$(&$s::LINT),*];
        }
    }
}
lints! {
    empty_label,
    relative_workdir,
    run_cd,
}

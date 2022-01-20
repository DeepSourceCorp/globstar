mod empty_label;
mod relative_workdir;
mod run_cd;

use aspen::Lint;

pub fn lints() -> Vec<&'static Lint> {
    vec![&empty_label::LINT, &relative_workdir::LINT, &run_cd::LINT]
}

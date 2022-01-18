use aspen::Lint;

mod assignment_in_condition;

pub fn lints() -> Vec<&'static Lint> {
    vec![&assignment_in_condition::LINT]
}

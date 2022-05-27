mod command_instead_of_module;
mod deprecated_local_action;
mod literal_compare;
mod partial_become;
mod variable_shadowing;
mod vcs_latest;

pub const LINTS: &[aspen::ValidatorFn] = &[
    variable_shadowing::validate,
    partial_become::validate,
    deprecated_local_action::validate,
    vcs_latest::validate,
    literal_compare::validate,
    command_instead_of_module::validate,
];

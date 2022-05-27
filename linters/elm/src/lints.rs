use aspen::{Lint, ValidatorFn};
mod variable_shadowing;

const VARIABLE_SHADOWING: Lint = Lint {
    name: "variable shadowing",
    code: "ELM-W1000",
};

pub const LINTS: [ValidatorFn; 1] = [variable_shadowing::validate];

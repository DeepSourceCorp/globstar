mod assignment_in_condition;
mod empty_to_json;
mod sussign;
mod useless_cmp;
mod variable_shadowing;

pub const LINTS: &[aspen::ValidatorFn] = &[
    assignment_in_condition::validate,
    useless_cmp::validate,
    sussign::validate,
    empty_to_json::validate,
    variable_shadowing::validate,
];

mod defs {
    use aspen::Lint;

    pub const ASSIGNMENT_IN_CONDITION: Lint = Lint {
        name: "assign-instead-of-eq",
        code: "RB-W1000",
    };

    pub const USELESS_CMP: Lint = Lint {
        name: "useless-cmp",
        code: "RB-W1001",
    };

    pub const SUSSIGN: Lint = Lint {
        name: "sussign",
        code: "RB-W1002",
    };

    pub const EMPTY_TO_JSON: Lint = Lint {
        name: "empty-to-json",
        code: "RB-W1003",
    };

    pub const VARIABLE_SHADOWING: Lint = Lint {
        name: "variable-shadowing",
        code: "RB-W1004",
    };
}

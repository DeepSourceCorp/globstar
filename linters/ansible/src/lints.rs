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

mod defs {
    use aspen::Lint;

    pub const VARIABLE_SHADOWING: Lint = Lint {
        name: "variable-shadowing",
        code: "YML-W1000",
    };

    pub const PARTIAL_BECOME: Lint = Lint {
        name: "partial-become",
        code: "YML-W1001",
    };

    pub const DEPRECATED_LOCAL_ACTION: Lint = Lint {
        name: "deprecated-local-action",
        code: "YML-W1002",
    };

    pub const GIT_LATEST: Lint = Lint {
        name: "git-latest",
        code: "YML-W1003",
    };

    pub const HG_LATEST: Lint = Lint {
        name: "hg-latest",
        code: "YML-W1004",
    };

    pub const LITERAL_COMPARE: Lint = Lint {
        name: "literal-compare",
        code: "YML-W1005",
    };

    pub const COMMAND_INSTEAD_OF_MODULE: Lint = Lint {
        name: "command-instead-of-module",
        code: "YML-W1006",
    };
}

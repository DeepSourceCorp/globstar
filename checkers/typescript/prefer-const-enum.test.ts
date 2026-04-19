// Test file for prefer-const-enum checker

//Should be flagged - regular enum
enum Color {
  Red = "RED",
  Green = "GREEN",
  Blue = "BLUE",
}

//Should be flagged - regular enum with numbers
enum Status {
  Active,
  Inactive,
  Pending,
}

//Should be flagged - regular enum
enum Direction {
  Up = 1,
  Down = 2,
  Left = 3,
  Right = 4,
}

//Should NOT be flagged - const enum
const enum Priority {
  Low = 1,
  Medium = 2,
  High = 3,
}

//Should NOT be flagged - const enum with strings
const enum Theme {
  Light = "LIGHT",
  Dark = "DARK",
}

//Should NOT be flagged - using union type instead
type ColorType = "RED" | "GREEN" | "BLUE";

const myColor: ColorType = "RED";

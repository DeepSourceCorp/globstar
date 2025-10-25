// Test file for ts-unused-import checker

// Should be flagged - unused import
import { unusedFunction } from "./utils";
import { anotherUnused } from "library";

// hould NOT be flagged - used import
import { usedFunction } from "./helpers";
import React from "react";

interface Props {
  value: string;
}

// Using React
const Component: React.FC<Props> = ({ value }) => {
  // Using usedFunction
  return usedFunction(value);
};

export default Component;

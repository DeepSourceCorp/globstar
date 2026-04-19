// Test file for ts-debugger-in-prod checker

function processData(data: number[]) {
  // <expect-error>
  debugger;
  return data.map((x) => x * 2);
}

class DataProcessor {
  process(items: string[]) {
    // <expect-error>
    debugger;
    return items.filter((x) => x.length > 0);
  }
}

const handleClick = () => {
  // <expect-error>
  debugger;
  console.log("clicked");
};

// ✅ Should NOT be flagged - no debugger
function processDataSafe(data: number[]) {
  console.log("Processing data:", data);
  return data.map((x) => x * 2);
}


interface Person {
    name: string;
    age: number;
}
// <expect-error>
const person = { name: "Alice", age: 30, extra: true } as Person;

// <no-error>
const person : Person = { name: "Alice", age: 30, extra: true };
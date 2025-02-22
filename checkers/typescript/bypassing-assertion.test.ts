// <expect-error>
interface Person {
    name: string;
    age: number;
}
const person = { name: "Alice", age: 30, extra: true } as Person;
// <no-error>
interface Person {
    name: string;
    age: number;
}
const person : Person = { name: "Alice", age: 30, extra: true };
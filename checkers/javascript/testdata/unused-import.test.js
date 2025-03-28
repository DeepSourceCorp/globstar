// <expect-error> unused-import: unused import "unused"
import { unused } from './module';

// Should not report used import
import { used } from './module';
console.log(used);

// Should not report used import
import * as namespace from 'module';

function test() {
    console.log(namespace);
}
// Match
// Source https://github.com/ethers-io/ethers.js/blob/main/src.ts/_admin/test-browser.ts#L290
import { createServer, Server } from "http";
// <expect-error>
const server = createServer((req, resp) => {});
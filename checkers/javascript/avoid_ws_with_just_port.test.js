// <expect-error>
const wss1 = new WebSocketServer({ port: 8080 });

import { createServer } from "https";
const server = createServer({ cert: ..., key: ... }); 
// <no-error>
const wss4 = new WebSocketServer({ server }); 
// <expect-error>
const wss2 = new WebSocketServer("ws://google.com");
// <no-error>
const wss3 = new WebSocketServer("wss://google.com");
// match
// <expect-error>
wss.on('connection', (ws : any, req: any) => {
    console.log('Browser connected to WebSocket server!');
})  
// Dont Match
// <no-errror>
wss.on('connection', (ws : any, req: any) => {
    const origin = req.headers.origin;
    if (origin !== "http://localhost:3000" && origin !== "https://www.youtube.com") {
        ws.close(); 
        return;
    }
})
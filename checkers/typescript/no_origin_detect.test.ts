// match
// <expect-error>
wss.on('connection', (ws : any, req: any) => {
    console.log('Browser connected to WebSocket server!');
    const origin = req.headers.origin;

    console.log(`Origin: ${origin}`);
    
})  
// Dont Match
// <no-errror>
wss.on('connection', (ws : any, req: any) => {
    console.log('Browser connected to WebSocket server!');

    const origin = req.headers.origin;

    console.log(`Origin: ${origin}`);
    
    if (origin !== "http://localhost:3000" && origin !== "https://www.youtube.com") {
        console.log(`Connection rejected from origin: ${origin}`);
        ws.close(); 
        return;
    }
})
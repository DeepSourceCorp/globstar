import express from 'express'
const app = express();
let app = express();
app.use('/api/:chain/:network', (req: Request, resp: Response, next: any) => {
  let { chain, network } = req.params;
 
  if (chain && !hasChain) {
    // <expect-error>
    return resp.status(500).send(`This node is not configured for the chain ${chain}`);
    // <expect-error>
    return resp.send(`This node is not configured for the chain ${chain}`);
  }
  return next();
  // <expect-error>
  return resp.send(`This node is not configured for the chain ${chain}`);
});

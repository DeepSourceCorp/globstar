let jwt = require('jsonwebtoken')
// <expect-error>
const dev = jwt.verify(token,null,{})
// <no-error>
const dec = jwt.verify(token,"HELLO")
const { Client } = require('ssh2');

const conn = new Client();

conn.on('ready', () => {
  console.log('Client :: ready');
  conn.end();
}).connect({
  host: 'example.com',
  port: 22,
  username: 'user',
  password: 'password',
  // <expect-error>
  hostVerifier: () => true // This blindly accepts any host key!
});



const trustedFingerprint = 'SHA256:your_known_fingerprint_here';

conn.connect({
  host: "example.com",
  port: 22,
  username: "user",
  password: "password",
  // <no-error>
  hostVerifier: (key) => {
    const fingerprint = crypto.createHash('sha256').update(key).digest('base64');
    return fingerprint === trustedFingerprint.split('SHA256:')[1];
  }
});

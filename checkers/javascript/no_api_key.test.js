
// Source : https://github.com/puppeteer/puppeteer/blob/main/website/docusaurus.config.js#L283
// Match
// <expect-error>
{apiKey: '4dac1ae64b623f1d33ae0b4ce0ff16a4',}
// <expect-error>
{APIKEY: '4dac1ae64b623f1d33ae0b4ce0ff16a4',}
// <expect-error>
algolia : {apiKey: '4dac1ae64b623f1d33ae0b4ce0ff16a4',}
// <expect-error>
algolia : {APIKEY: '4dac1ae64b623f1d33ae0b4ce0ff16a4',}

// Dont Match
// <no-error>
{KAPIKEY: '4dac1ae64b623f1d33ae0b4ce0ff16a4',}
// <no-error>
{APIKEYY: '4dac1ae64b623f1d33ae0b4ce0ff16a4',}
// <no-error>
algolia : {APIKEYY: '4dac1ae64b623f1d33ae0b4ce0ff16a4',}

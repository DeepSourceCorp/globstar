// macth
// <expext-error>
res.cookie('user', 'john_doe', { maxAge: 900000, httpOnly: false });
// <expext-error>
res.cookie('user', 'john_doe', { httpOnly: false });
// <expext-error>
res.cookie('user', 'john_doe', { httpOnly: false , maxAge: 900000, });
// <expext-error>
res.cookie('user', 'john_doe', { secure : false });
// <expext-error>
res.cookie('user', 'john_doe', { sameSite : false });
// <expext-error>
res.cookie('user', 100, { sameSite : false });
// <expext-error>
res.cookie('user', {my_name : "some_name"}, { sameSite : false });

// Dont Match
// <no-error>
res.cookie('user', 'john_doe', { secure : true });
// <no-error>
res.cookie('user', 'john_doe', { httpOnly: true });
// <no-error>
res.cookie('user', 'john_doe', { sameSite : Strict });
// <no-error>
res.cookie('user', {my_name : "some_name"}, { sameSite : true });
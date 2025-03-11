// <expect-error>
http.SetCookie(&http.Cookie{
    Name: "session",
    Value: "123456",
	Secure: false,
})

// <expect-error>
http.SetCookie(&http.Cookie{
	Name: "session",
	Value: "123456",
})

// Safe - setting Secure to true
http.SetCookie(&http.Cookie{
	Name: "session",
	Value: "123456",
	Secure: true
})
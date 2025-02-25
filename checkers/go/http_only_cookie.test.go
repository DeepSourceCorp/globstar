// <expect-error> httpOnly is not set
http.SetCookie(&http.Cookie{
    Name: "session",
    Value: "123456",
})

// <expect-error> httpOnly is set to false
http.SetCookie(&http.Cookie{
    Name: "session",
    Value: "123456",
    HttpOnly: false
})

// Safe - setting HttpOnly to true
http.SetCookie(&http.Cookie{
    Name: "session",
    Value: "123456",
    HttpOnly: true 
})
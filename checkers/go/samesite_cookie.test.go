import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Insecure: SameSite=None without Secure flag (CSRF vulnerability)
	// <expect-error>
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "123456",
		SameSite: http.SameSiteNoneMode,
	})

	// Insecure: Missing SameSite attribute (defaults to SameSiteDefaultMode, can be risky)
	// <expect-error>
	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: "123456",
	})

	// Secure: SameSite set to LaxMode (recommended for most cases)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "123456",
		SameSite: http.SameSiteLaxMode,
	})

	// Secure: SameSite set to StrictMode (maximum CSRF protection)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_strict",
		Value:    "7891011",
		SameSite: http.SameSiteStrictMode,
	})

	fmt.Fprintln(w, "Cookies set!")
}

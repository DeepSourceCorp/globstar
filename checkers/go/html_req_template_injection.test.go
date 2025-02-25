import (
	"fmt"
	"net/http"
	"text/template"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// <expect-error> Unsanitized user input from request form passed to template
	userInput := r.FormValue("name")

	tmpl := template.New("mytemplate")

	parsedTmpl, err := tmpl.Parse("Hello, {{.Name}}!")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	data := map[string]string{"Name": userInput}
	err = parsedTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// <expect-error> Unsanitized user input from request cookie passed to template
	userInput := r.Cookie("session_id")

	tmpl := template.New("mytemplate")

	parsedTmpl, err := tmpl.Parse("Hello, {{.Name}}!")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	data := map[string]string{"Name": userInput}
	err = parsedTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}


func handler(w http.ResponseWriter, r *http.Request) {
	// <expect-error> Unsanitized user input from request query params passed to template
	userInput := r.URL.Query().Get("name")

	tmpl := template.New("mytemplate")

	parsedTmpl, err := tmpl.Parse("Hello, {{.Name}}!")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	data := map[string]string{"Name": userInput}
	err = parsedTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}



func handler(w http.ResponseWriter, r *http.Request) {
	userInput := r.FormValue("name")

	tmpl := template.New("mytemplate")

	parsedTmpl, err := tmpl.Parse("Hello, {{.Name}}!")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	// Sanitize user input
	SanitizedInput := template.HTMLEscapeString(userInput)

	data := map[string]string{"Name": SanitizedInput}
	// Safe - Data contains sanitized user input
	err = parsedTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}
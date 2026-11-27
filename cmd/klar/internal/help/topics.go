package help

type Topic struct {
	Title       string
	URL         string
	Description string
	SeeAlso     []string
}

var Topics = map[string]*Topic{
	"js": {
		Title:       "JavaScript Compilation",
		Description: "Hello, JavaScript Compiler!",
		URL:         "https://klarlanguage.github.io/docs/js",
		SeeAlso:     nil,
	},
}

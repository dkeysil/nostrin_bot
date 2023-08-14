package messages

import "text/template"

var (
	NoteTextTemplate = template.Must(
		template.New("message").Parse(`<b>{{.DisplayName}} {{if .Nip05 }}({{.Nip05}}){{end}}</b>
<code>
{{.Text}}
</code>
{{.RepliesCount}} Replies {{.RepostsCount}} Reposts {{.LikesCount}} Likes {{.ZapsSum}} Zaps
<a href="{{.Link}}">Open on nostter.com</a>		{{.CreatedAt}}
`),
	)
)

type NoteTemplate struct {
	DisplayName  string
	Nip05        string
	Text         string
	CreatedAt    string
	Link         string
	RepliesCount int
	RepostsCount int
	LikesCount   int
	ZapsSum      int64
}

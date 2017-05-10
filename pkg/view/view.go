package view

import "net/http"

type viewKey int

const (
	_ viewKey = iota
	keyIndex
	keyProfile
	keyProfileEdit
	keyUser
	keyCourse
	keyCourseEdit
)

// Index renders index view
func Index(w http.ResponseWriter, r *http.Request, data *IndexData) {
	render(w, r, keyIndex, data)
}

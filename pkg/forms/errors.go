package forms

type errors map[string][]string

func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}
func (e errors) Get(field string) string {
	es, ok := e[field]
	if !ok {
		return ""
	}
	return es[0]
}

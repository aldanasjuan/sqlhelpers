package sqlhelpers

type Set map[string]struct{}

func (s *Set) Add(st string) {
	v := *s
	v[st] = struct{}{}
}
func (s *Set) Remove(st string) {
	delete(*s, st)
}
func (s *Set) Exists(st string) bool {
	v := *s
	_, ok := v[st]
	return ok
}

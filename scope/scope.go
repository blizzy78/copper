package scope

// Scope is a map of values indexed by identifiers.
//
// A scope may have a parent scope. If the current scope does not store a value for a
// specific identifier, the parent scopes will be considered (recursively.)
type Scope struct {
	// Parent is the parent scope of this scope.
	Parent *Scope

	values map[string]interface{}
	locked bool
}

// Set stores the value v identified by name in the scope.
//
// If the scope has a parent scope, and that parent scope already stores a value for that
// identifier, the value is stored (overwritten) in the parent scope instead. Parent scopes are considered
// recursively until there is no parent scope.
//
// If the scope where the value should be stored is locked, nothing will happen.
func (s *Scope) Set(name string, v interface{}) {
	ps := s.Parent
	parentHasValue := false
	for ps != nil {
		if parentHasValue = hasValueSelf(ps, name); parentHasValue {
			break
		}

		ps = ps.Parent
	}

	if parentHasValue {
		ps.values[name] = v
		return
	}

	if s.locked {
		return
	}

	s.init()
	s.values[name] = v
}

// HasValue returns whether the scope or any of its parent scopes store a value identified by name.
func (s *Scope) HasValue(name string) (ok bool) {
	for {
		if ok = hasValueSelf(s, name); ok {
			break
		}

		if s = s.Parent; s == nil {
			break
		}
	}

	return
}

// Value returns the value identified by name in this scope or any of its parent scopes.
// If there is a value, ok will be true, otherwise it will be false.
func (s *Scope) Value(name string) (v interface{}, ok bool) {
	for {
		if s.values != nil {
			if v, ok = s.values[name]; ok {
				break
			}
		}

		if s = s.Parent; s == nil {
			break
		}
	}

	return
}

// Lock prevents this scope from further modification. Parent scopes (if any) will not be locked.
func (s *Scope) Lock() {
	s.locked = true
}

// ClearSelf removes all values associated with this scope, not including any parent scopes.
func (s *Scope) ClearSelf() {
	if s.values != nil {
		for k := range s.values {
			delete(s.values, k)
		}
	}
}

func (s *Scope) init() {
	if s.values == nil {
		s.values = map[string]interface{}{}
	}
}

func hasValueSelf(s *Scope, name string) (ok bool) {
	if s.values != nil {
		_, ok = s.values[name]
	}
	return
}

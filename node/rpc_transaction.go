package node

// Args is used to define method arguments
type Args struct {
	A, B int
}

// Result is the result of a method call
type Result map[string]interface{}

// Send creates a
func (s *Service) Send(args Args, result *Result) error {
	r := map[string]interface{}{
		"sum": args.A + args.B,
	}
	*result = r
	return nil
}

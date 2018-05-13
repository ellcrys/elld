package rpc

// Args is used to define method arguments
type Args struct {
	A, B int
}

// Send creates a
func (s *Service) Send(args Args, result *Result) error {
	// r := map[string]interface{}{
	// 	"sum": args.A + args.B,
	// }
	// *result = r
	return nil
}

package runtime

import "fmt"

// DeniedError is returned by Evaluate when the policy engine denies a tool call.
// Agent code can use errors.As to distinguish a policy denial from a real tool failure.
type DeniedError struct {
	Tool        string
	Reason      string
	RuleMatched string
}

func (e *DeniedError) Error() string {
	return fmt.Sprintf("tool %q denied by policy: %s", e.Tool, e.Reason)
}

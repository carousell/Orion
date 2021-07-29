package forwardheaders

// AllowForwardHeaders is an interface that you need to implement for Orion's services to set the allowlist in the ctx
type AllowForwardHeaders interface {
	GetAllowedForwardHeaders() []string
}

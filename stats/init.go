package stats

var (
	root *registry
	// Root is a reference the Registry singleton.
	Root Registry
)

// Ensure that container is always defined and contains a "statistics" map.
func init() {
	root = newRegistry()
	Root = root
}

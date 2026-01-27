//go:build android || darwin || ios

package widgets

import "github.com/go-drift/drift/pkg/semantics"

func (r *renderSvgImage) DescribeSemanticsConfiguration(config *semantics.SemanticsConfiguration) bool {
	if r.excludeFromSemantics {
		return false
	}

	config.Properties.Role = semantics.SemanticsRoleImage
	config.Properties.Flags = config.Properties.Flags.Set(semantics.SemanticsIsImage)

	if r.semanticLabel != "" {
		config.Properties.Label = r.semanticLabel
	}

	// Note: source field is not used in semantics, only semanticLabel matters
	return true
}

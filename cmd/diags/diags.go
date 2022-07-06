package diags

import (
	"sort"

	"github.com/cloudquery/cloudquery/pkg/ui"
	"github.com/cloudquery/cq-provider-sdk/provider/diag"
)

func PrintDiagnostics(header string, dd *diag.Diagnostics, redactDiags, verbose bool) {
	// Nothing to
	if dd == nil || !dd.HasDiags() {
		return
	}
	diags := *dd

	if redactDiags {
		diags = diags.Redacted()
	}

	diags = diags.Squash()

	// classify diags to user diags + add details best on received error messages
	diags = classifyDiagnostics(diags)

	if !verbose {
		var hasPrintableDiag bool
		for _, d := range diags {
			if d.Severity() != diag.IGNORE {
				hasPrintableDiag = true
				break
			}
		}
		if !hasPrintableDiag {
			return
		}
	}

	// sort diagnostics by severity/type
	sort.Sort(diags)

	if header != "" {
		ui.ColorizedNoLogOutput(ui.ColorHeader, "%s Diagnostics:\n\n", header)
	} else {
		ui.ColorizedNoLogOutput(ui.ColorHeader, "Diagnostics:\n\n")
	}

	for _, d := range diags {
		if !verbose && d.Severity() == diag.IGNORE {
			continue
		}
		printDiagnostic(d)
	}
	ui.ColorizedOutput(ui.ColorInfo, "\n")
}

package presentation

import (
	"io"

	"github.com/manisbindra/az-mpf/pkg/domain"
)

type DisplayOptions struct {
	ShowDetailedOutput             bool
	JSONOutput                     bool
	DefaultResourceGroupResourceID string
}

// type ResultDisplayer interface {
// 	DisplayMPFResult(w io.Writer ,result domain.MPFResult, options displayOptions) error
// }

type displayConfig struct {
	result         domain.MPFResult
	displayOptions DisplayOptions
}

func NewMPFResultDisplayer(result domain.MPFResult, options DisplayOptions) *displayConfig {
	return &displayConfig{
		result:         result,
		displayOptions: options,
	}
}

func (d *displayConfig) DisplayResult(w io.Writer) error {
	if d.displayOptions.JSONOutput {
		return d.displayJSON(w)
	}
	return d.displayText(w)
}

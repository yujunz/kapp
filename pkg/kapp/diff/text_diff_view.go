package diff

import (
	"fmt"
	"strings"

	"github.com/aryann/difflib"
	"github.com/fatih/color"
	ctlconf "github.com/k14s/kapp/pkg/kapp/config"
)

type TextDiffViewOpts struct {
	Context int // number of lines to show around changed lines; <0 for all
	Mask    bool
}

type TextDiffView struct {
	diff      *ConfigurableTextDiff
	maskRules []ctlconf.DiffMaskRule
	opts      TextDiffViewOpts
}

func NewTextDiffView(diff *ConfigurableTextDiff,
	maskRules []ctlconf.DiffMaskRule, opts TextDiffViewOpts) TextDiffView {

	return TextDiffView{diff, maskRules, opts}
}

func (v TextDiffView) String() string {
	var diffRecords []difflib.DiffRecord

	if v.opts.Mask {
		textDiff, err := v.diff.Masked(v.maskRules)
		if err != nil {
			return fmt.Sprintf("Error masking diff: %s", err)
		}
		diffRecords = textDiff.Records()
	} else {
		diffRecords = v.diff.Full().Records()
	}

	lines := []string{}
	changedLines := map[int]struct{}{}

	for lineNum, diff := range diffRecords {
		if diff.Delta != difflib.Common {
			changedLines[lineNum] = struct{}{}
		}
	}

	prevInContext := false
	emptyLineStr := "   "
	lineStr := func(line int) string { return fmt.Sprintf("%3d", line) }

	for lineNum, diff := range diffRecords {
		switch diff.Delta {
		case difflib.RightOnly:
			lines = append(lines, color.New(color.FgGreen).Sprintf("%s %s + %s",
				emptyLineStr,
				lineStr(diff.LineRight),
				diff.Payload))

		case difflib.LeftOnly:
			lines = append(lines, color.New(color.FgRed).Sprintf("%s %s - %s",
				lineStr(diff.LineLeft),
				emptyLineStr,
				diff.Payload))

		case difflib.Common:
			newInContext := v.inContext(lineNum, changedLines)
			if lineNum != 0 && !prevInContext && newInContext {
				lines = append(lines, "  ...")
			}
			if newInContext {
				lines = append(lines, fmt.Sprintf("%s,%s   %s",
					lineStr(diff.LineLeft),
					lineStr(diff.LineRight),
					diff.Payload)) // LineLeft == LineRight
			}
			prevInContext = newInContext
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

func (v TextDiffView) inContext(lineNum int, changedLines map[int]struct{}) bool {
	if v.opts.Context < 0 {
		return true
	}
	for i := lineNum - v.opts.Context; i < lineNum+v.opts.Context; i++ {
		if _, found := changedLines[i]; found {
			return true
		}
	}
	return false
}

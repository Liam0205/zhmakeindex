package index

import (
	"fmt"
	"io"

	"github.com/leo-liu/zhmakeindex/internal/page"
	"github.com/leo-liu/zhmakeindex/internal/style"
)

type PageRange struct {
	Begin *page.Page
	End   *page.Page
}

func (p *PageRange) Diff() int {
	return p.End.Diff(p.Begin)
}

func (p *PageRange) Write(out io.Writer, st *style.OutputStyle) {
	var rangestr string
	switch {
	case p.Diff() == 0:
		rangestr = p.Begin.String()
	case p.Begin.Rangetype == page.PAGE_NORMAL && p.End.Rangetype == page.PAGE_NORMAL &&
		p.Diff() == 1 && st.Suffix2p == "":
		rangestr = p.Begin.String() + st.DelimN + p.End.String()
	case p.Diff() == 1 && st.Suffix2p != "":
		rangestr = p.Begin.String() + st.Suffix2p
	case p.Diff() == 2 && st.Suffix3p != "":
		rangestr = p.Begin.String() + st.Suffix3p
	case p.Diff() >= 2 && st.SuffixMp != "":
		rangestr = p.Begin.String() + st.SuffixMp
	default:
		rangestr = p.Begin.String() + st.DelimR + p.End.String()
	}
	if p.Begin.Encap == "" {
		fmt.Fprint(out, rangestr)
	} else {
		fmt.Fprint(out, st.EncapPrefix, p.Begin.Encap,
			st.EncapInfix, rangestr, st.EncapSuffix)
	}
}

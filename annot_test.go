package annot

import (
	"bytes"
	"testing"
)

func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		annots  []*Annot
		wantW   string
		wantErr bool
	}{
		{
			name: "empty annotation",
			annots: []*Annot{
				{},
			},
			wantW: `
↑
└─ 
`,
			wantErr: false,
		},
		{
			name: "two empty annotations",
			annots: []*Annot{
				{Col: 0},
				{Col: 2},
			},
			wantW: `
↑ ↑
│ └─ 
│
└─ 
`,
			wantErr: false,
		},
		{
			name: "annotation without a column",
			annots: []*Annot{
				{Lines: []string{"line1"}},
			},
			wantW: `
↑
└─ line1
`,
			wantErr: false,
		},
		{
			name: "annotation without a line",
			annots: []*Annot{
				{Col: 0},
			},
			wantW: `
↑
└─ 
`,
			wantErr: false,
		},
		{
			name: "annotation with one line",
			annots: []*Annot{
				{Col: 1, Lines: []string{"line1"}},
			},
			wantW: `
 ↑
 └─ line1
`,
			wantErr: false,
		},
		{
			name: "annotation with five lines",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1", "line2", "line3", "line4", "line5"}},
			},
			wantW: `
↑
└─ line1
   line2
   line3
   line4
   line5
`,
			wantErr: false,
		},
		{
			name: "utf8, japanese characters and emojis",
			annots: []*Annot{
				{Col: 0, Lines: []string{"⭐️漢", "æñ🥏Ǌ"}},
				{Col: 10, Lines: []string{"æñŶǼǊ", "字ñŶǼǊ"}},
			},
			wantW: `
↑         ↑
└─ ⭐️漢   └─ æñŶǼǊ
   æñ🥏Ǌ     字ñŶǼǊ
`,
			wantErr: false,
		},
		{
			name: "next to each other with enough distance and second annotation has more lines",
			annots: []*Annot{
				{Col: 5, Lines: []string{"line1", "line2"}},
				{Col: 20, Lines: []string{"line1", "line2", "line3", "line4"}},
			},
			wantW: `
     ↑              ↑
     └─ line1       └─ line1
        line2          line2
                       line3
                       line4
`,
			wantErr: false,
		},
		{
			name: "next to each other with enough distance and first annotation has more lines",
			annots: []*Annot{
				{Col: 5, Lines: []string{"line1", "line2", "line3", "line4"}},
				{Col: 20, Lines: []string{"line1", "line2"}},
			},
			wantW: `
     ↑              ↑
     └─ line1       └─ line1
        line2          line2
        line3
        line4
`,
			wantErr: false,
		},
		{
			name: "annots are close",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1", "line2"}},
				{Col: 1, Lines: []string{"line1", "line2"}},
			},
			wantW: `
↑↑
│└─ line1
│   line2
│
└─ line1
   line2
`,
			wantErr: false,
		},
		{
			name: "annots have correct vertical distance",
			annots: []*Annot{
				{Col: 5, Lines: []string{"line1", "line2"}},
				{Col: 10, Lines: []string{"line1", "line2"}},
				{Col: 15, Lines: []string{"line1", "line2"}},
			},
			wantW: `
     ↑    ↑    ↑
     │    │    └─ line1
     │    │       line2
     │    │
     │    └─ line1
     │       line2
     │
     └─ line1
        line2
`,
			wantErr: false,
		},
		{
			name: "first annotation is at the height of the pipes of the second annotation",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1", "line2"}},
				{Col: 10, Lines: []string{"line1", "line2"}},
				{Col: 15, Lines: []string{"line1", "line2"}},
			},
			wantW: `
↑         ↑    ↑
└─ line1  │    └─ line1
   line2  │       line2
          │
          └─ line1
             line2
`,
			wantErr: false,
		},
		{
			name: "correct vertical and horizontal distance",
			annots: []*Annot{
				{Col: 5, Lines: []string{"line1", "line2"}},
				{Col: 10, Lines: []string{"line1", "line2"}},
				{Col: 20, Lines: []string{"line1", "line2"}},
			},
			wantW: `
     ↑    ↑         ↑
     │    └─ line1  └─ line1
     │       line2     line2
     │
     └─ line1
        line2
`,
			wantErr: false,
		},
		{
			name: "correct horizontal distance",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1", "line2"}},
				{Col: 10, Lines: []string{"line1", "line2"}},
			},
			wantW: `
↑         ↑
└─ line1  └─ line1
   line2     line2
`,
			wantErr: false,
		},
		{
			name: "long line of first annotation uses available space",
			annots: []*Annot{
				{Col: 2, Lines: []string{"line1", "line2", "line3long"}},
				{Col: 10, Lines: []string{"line1", "line2", "line3"}},
			},
			wantW: `
  ↑       ↑
  │       └─ line1
  │          line2
  └─ line1   line3
     line2
     line3long
`,
			wantErr: false,
		},
		{
			name: "last annotation shares row with first annotation",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1", "line2"}},
				{Col: 10, Lines: []string{"line1"}},
				{Col: 20, Lines: []string{"line1", "line2"}},
			},
			wantW: `
↑         ↑         ↑
└─ line1  └─ line1  └─ line1
   line2               line2
`,
			wantErr: false,
		},
		{
			name: "complex annotation arrangement",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1", "line2", "line3", "line4", "line5", "line6", "line7", "line8verylongverylongverylong"}},
				{Col: 10, Lines: []string{"line1"}},
				{Col: 20, Lines: []string{"line1"}},
				{Col: 25, Lines: []string{"line1", "line2"}},
				{Col: 35, Lines: []string{"line1long", "line2", "line3"}},
				{Col: 40, Lines: []string{"line1", "line2", "line3", "line4"}},
			},
			wantW: `
↑         ↑         ↑    ↑         ↑    ↑
└─ line1  └─ line1  │    └─ line1  │    └─ line1
   line2            │       line2  │       line2
   line3            │              │       line3
   line4            └─ line1       │       line4
   line5                           │
   line6                           └─ line1long
   line7                              line2
   line8verylongverylongverylong      line3
`,
			wantErr: false,
		},
		{
			name: "last annotation shares row with first annotation and first annotation uses available space",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1", "line2", "line3verylongverylong"}},
				{Col: 10, Lines: []string{"line1"}},
				{Col: 20, Lines: []string{"line1", "line2", "line3"}},
			},
			wantW: `
↑         ↑         ↑
│         └─ line1  └─ line1
│                      line2
└─ line1               line3
   line2
   line3verylongverylong
`,
			wantErr: false,
		},
		{
			name: "first annotation uses indentation space of second annotation",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1"}},
				{Col: 8, Lines: []string{"line1", "line2", "line3"}},
			},
			wantW: `
↑       ↑
│       └─ line1
│          line2
└─ line1   line3
`,
			wantErr: false,
		},
		{
			name: "remove second annotation with same column position",
			annots: []*Annot{
				{Col: 0, Lines: []string{"line1"}},
				{Col: 0, Lines: []string{"same column position"}},
			},
			wantW: `
↑
└─ line1
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := Write(w, tt.annots...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := "\n" + w.String(); gotW != tt.wantW {
				t.Errorf("Write() gotW = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

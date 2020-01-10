package draw

import (
	"github.com/gdamore/tcell"
)

const (
	// TimerWidth is width timer
	TimerWidth = 21
	// TimerHeight is height timer
	TimerHeight = 5

	numberWidth  = 4
	numberHeight = 5

	separaterWidth  = 1
	separaterHeight = 5

	whitespaceWidth = 1

	// StatusBarBackgroundColor status bar background color
	StatusBarBackgroundColor = tcell.ColorBlack
)

const separator = `
-
#
-
#
-
`

var numbers = []string{
	`
####
#--#
#--#
#--#
####
	`,
	`
---#
---#
---#
---#
---#
`,
	`
####
---#
####
#---
####
`,
	`
####
---#
####
---#
####
`,
	`
#-#-
#-#-
####
--#-
--#-
`,
	`
####
#---
####
---#
####
`,
	`
#---
#---
####
#--#
####
`,
	`
####
---#
---#
---#
---#
`,
	`
####
#--#
####
#--#
####
`,
	`
####
#--#
####
---#
####
`,
}

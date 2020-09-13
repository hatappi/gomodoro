package draw

import (
	"github.com/gdamore/tcell"
)

const (
	// TimerBaseWidth is base width of timer
	TimerBaseWidth = 21
	// TimerBaseHeight is base height of timer
	TimerBaseHeight = 5

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

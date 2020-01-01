package screen

const (
	TIMER_WIDTH  = 21
	TIMER_HEIGHT = 5

	number_width  = 4
	number_height = 5

	separater_width  = 1
	separater_height = 5

	whitespace_width = 1
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

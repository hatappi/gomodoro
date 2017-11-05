package drawer

import (
	"strings"
)

var separator = string(`
-----
--#--
-----
--#--
-----
`)

var strNumbers = []string{
	`
#####
#---#
#---#
#---#
#####
`,
	`
----#
----#
----#
----#
----#
`,
	`
#####
----#
#####
#----
#####
`,
	`
#####
----#
#####
----#
#####
`,
	`
#--#-
#--#-
#####
---#-
---#-
`,
	`
#####
#----
#####
----#
#####
`,
	`
#----
#----
#####
#---#
#####
`,
	`
#####
----#
----#
----#
----#
`,
	`
#####
#---#
#####
#---#
#####
`,
	`
#####
#---#
#####
----#
#####
`,
}

func Num2StrArray(n int) []string {
	return strings.Split(strNumbers[n], "")
}

func SeparatorStrArray() []string {
	return strings.Split(separator, "")
}

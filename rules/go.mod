module gobox/rules

go 1.26.2

require gobox/char v0.0.0

replace gobox/char => ../char

require (
	gobox/core v0.0.0
	rules1e v0.0.0
)

replace gobox/core => ../core

replace rules1e => ../../rules1e

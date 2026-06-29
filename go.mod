module github.com/EvilMilko/Obfuscator-1C

go 1.23.4

toolchain go1.24.1

//replace github.com/LazarenkoA/1c-language-parser => ..\1c-language-parser

require (
	github.com/LazarenkoA/1c-language-parser v0.0.0-20251013110129-461cd761a494
	github.com/google/uuid v1.6.0
	github.com/knetic/govaluate v3.0.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.53.0
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/text v0.22.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

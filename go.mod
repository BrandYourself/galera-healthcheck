module github.com/BrandYourself/galera-healthcheck

go 1.14

replace github.com/mingcheng/pidfile => github.com/mingcheng/pidfile.go v0.0.0-20190611121011-5e445891c73f

require (
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5
	github.com/go-sql-driver/mysql v1.5.0
	github.com/mingcheng/pidfile v0.0.0-00010101000000-000000000000
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/urfave/cli/v2 v2.2.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)

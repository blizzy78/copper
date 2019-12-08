module local/fuzz

go 1.13

replace github.com/blizzy78/copper => ../..

require (
	github.com/blizzy78/copper v0.0.0-00010101000000-000000000000
	github.com/dvyukov/go-fuzz v0.0.0-20191206100749-a378175e205c
	github.com/elazarl/go-bindata-assetfs v1.0.0 // indirect
	github.com/gobuffalo/nulls v0.1.0
	github.com/stephens2424/writerset v1.0.2 // indirect
)

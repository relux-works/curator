module example.test/cli

go 1.23

require (
	example.test/board v0.0.0
	example.test/remoteconfig v0.0.0
)

replace example.test/board => ../../pkg/board

replace example.test/remoteconfig => ../../pkg/remoteconfig

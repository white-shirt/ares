package logger

// Option logger Option func
type Option func(*Options)

// Options logger Options
type Options struct {
	directory     string // path为空，则使用directory + #{appname} + access.json 规则生成一个path
	path          string // 优先使用path
	consoleOutput bool
}

// Directory 相对于application.LogDir()之后的路径。
func Directory(directory string) Option {
	return func(opts *Options) {
		opts.directory = directory
	}
}

// ConsoleOutput 是否启用终端输出。
func ConsoleOutput(output bool) Option {
	return func(opts *Options) {
		opts.consoleOutput = output
	}
}

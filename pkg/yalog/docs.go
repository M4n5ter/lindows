/*
yalog 包封装了 slog 包，提供了更简单的接口。并且提供了一个全局的 logger(同时包含了使用[slog.TextHandler]和[slog.JSONHandler]的Logger)，可以直接使用。
package yalog encapsulates the slog package and provides a simpler interface. And provides a global logger (which contains both the Logger using [slog.TextHandler] and [slog.JSONHandler]), which can be used directly.

# Example

	yalog.SetLevelInfo()
	yalog.Debugf("hello %s", "world")
	yalog.Infof("hello %s", "world")
	yalog.Warnf("hello %s", "world")
	yalog.Errorf("hello world")
	yalog.Debug("hello world", "age", 18)
	yalog.Info("hello world", "age", 18)
	yalog.Warn("hello world", "age", 18)
	yalog.Error("hello world", "age", 18)

	l := yalog.Default()
	l.LogAttrs(context.Background(), yalog.LevelInfo, "hello world", yalog.Int("age", 22))
	l.Log(context.Background(), yalog.LevelInfo, "hello world", "age", 18)
	l.Debugf("hello %s", "world")
	l.Infof("hello %s", "world")
	l.Warnf("hello %s", "world")
	l.Errorf("hello world")
*/
package yalog

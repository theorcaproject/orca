package base


type AppName string
type AppType string
type AppStatus string


type AppConfiguration struct {
	Version Version
	Name AppName
	Type AppType
	InstallCommands []OsCommand
	QueryStateCommand OsCommand
	RemoveCommand OsCommand
}


package gcli

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gookit/color"
)

// core definition TODO rename to context ??
type core struct {
	*cmdLine
	SimplePrinter
	// HelpVars help template vars.
	HelpVars
	// Hooks manage. allowed hooks: "init", "before", "after", "error"
	Hooks
	// global options flag set
	gFlags *Flags
	// GOptsBinder you can custom binding global options
	GOptsBinder func(gf *Flags)
}

// init core
// func (c core) init(cmdName string) {
// 	c.cmdLine = CLI
//
// 	c.AddVars(c.innerHelpVars())
// 	c.AddVars(map[string]string{
// 		"cmd": cmdName,
// 		// binName with command
// 		"binWithCmd": c.binName + " " + cmdName,
// 		// binFile with command
// 		"fullCmd": c.binFile + " " + cmdName,
// 	})
// }

func (c core) parseGlobalOpts(args []string) (ok bool) {
	if c.gFlags == nil { // skip on nil
		return true
	}

	// parse global options
	err := c.gFlags.Parse(args)
	if err != nil {
		color.Error.Tips(err.Error())
		return
	}

	return true
}

// GlobalFlags get the app GlobalFlags
func (c core) GlobalFlags() *Flags {
	return c.gFlags
}

// common basic help vars
func (c core) innerHelpVars() map[string]string {
	return map[string]string{
		"pid":     CLI.PIDString(),
		"workDir": CLI.workDir,
		"binFile": CLI.binFile,
		"binName": CLI.binName,
	}
}

// SimplePrinter struct. for inject struct
type SimplePrinter struct {}

// Print message
func (s SimplePrinter) Print(v ...interface{}) {
	color.Print(v...)
}

// Printf message
func (s SimplePrinter) Printf(format string, v ...interface{}) {
	color.Printf(format, v...)
}

// Println message
func (s SimplePrinter) Println(v ...interface{}) {
	color.Println(v...)
}

// Infoln message
func (s SimplePrinter) Infoln(a ...interface{}) {
	color.Info.Println(a...)
}

// Warnln message
func (s SimplePrinter) Warnln(a ...interface{}) {
	color.Warn.Println(a...)
}

// Errorln message
func (s SimplePrinter) Errorln(a ...interface{}) {
	color.Error.Println(a...)
}

/*************************************************************
 * simple events manage
 *************************************************************/

// Hooks struct
type Hooks struct {
	// Hooks can setting some hooks func on running.
	hooks map[string]HookFunc
}

// On register event hook by name
func (h *Hooks) On(name string, handler HookFunc) {
	if handler != nil {
		if h.hooks == nil {
			h.hooks = make(map[string]HookFunc)
		}

		h.hooks[name] = handler
	}
}

// AddOn register on not exists hook.
func (h *Hooks) AddOn(name string, handler HookFunc) {
	if _, ok := h.hooks[name]; !ok {
		h.On(name, handler)
	}
}

// Fire event by name, allow with event data
func (h *Hooks) Fire(event string, data ...interface{}) {
	if handler, ok := h.hooks[event]; ok {
		handler(data...)
	}
}

// ClearHooks clear hooks data
func (h *Hooks) ClearHooks() {
	h.hooks = nil
}

/*************************************************************
 * Command Line: command data
 *************************************************************/

// cmdLine store common data for CLI
type cmdLine struct {
	// pid for current application
	pid int
	// os name.
	osName string
	// the CLI app work dir path. by `os.Getwd()`
	workDir string
	// bin script file, by `os.Args[0]`. eg "./path/to/cliapp"
	binFile string
	// bin script filename. eg "cliapp"
	binName string
	// os.Args to string, but no binName.
	argLine string
}

func newCmdLine() *cmdLine {
	binFile := os.Args[0]
	workDir, _ := os.Getwd()

	// binName will contains work dir path on windows
	// if envutil.IsWin() {
	// 	binFile = strings.Replace(CLI.binName, workDir+"\\", "", 1)
	// }

	return &cmdLine{
		pid: os.Getpid(),
		// more info
		osName:  runtime.GOOS,
		workDir: workDir,
		binFile: binFile,
		binName: filepath.Base(binFile),
		argLine: strings.Join(os.Args[1:], " "),
	}
}

// PID get pid
func (c *cmdLine) PID() int {
	return c.pid
}

// PIDString get pid as string
func (c *cmdLine) PIDString() string {
	return strconv.Itoa(c.pid)
}

// OsName is equals to `runtime.GOOS`
func (c *cmdLine) OsName() string {
	return c.osName
}

// OsArgs is equals to `os.Args`
func (c *cmdLine) OsArgs() []string {
	return os.Args
}

// BinName get bin script file
func (c *cmdLine) BinFile() string {
	return c.binFile
}

// BinName get bin script name
func (c *cmdLine) BinName() string {
	return c.binName
}

// BinDir get bin script dirname
func (c *cmdLine) BinDir() string {
	return path.Dir(c.binFile)
}

// WorkDir get work dirname
func (c *cmdLine) WorkDir() string {
	return c.workDir
}

// ArgLine os.Args to string, but no binName.
func (c *cmdLine) ArgLine() string {
	return c.argLine
}

func (c *cmdLine) hasHelpKeywords() bool {
	if c.argLine == "" {
		return false
	}

	return strings.HasSuffix(c.argLine, " -h") || strings.HasSuffix(c.argLine, " --help")
}

/*************************************************************
 * app/cmd help vars
 *************************************************************/

// HelpVarFormat allow var replace on render help info.
// Default support:
// 	"{$binName}" "{$cmd}" "{$fullCmd}" "{$workDir}"
const HelpVarFormat = "{$%s}"

// HelpVars struct. provide string var function for render help template.
type HelpVars struct {
	// varLeft, varRight string
	// varFormat string
	// Vars you can add some vars map for render help info
	Vars map[string]string
}

// AddVar get command name
func (hv *HelpVars) AddVar(name, value string) {
	if hv.Vars == nil {
		hv.Vars = make(map[string]string)
	}

	hv.Vars[name] = value
}

// AddVars add multi tpl vars
func (hv *HelpVars) AddVars(vars map[string]string) {
	for n, v := range vars {
		hv.AddVar(n, v)
	}
}

// GetVar get a help var by name
func (hv *HelpVars) GetVar(name string) string {
	return hv.Vars[name]
}

// GetVars get all tpl vars
func (hv *HelpVars) GetVars() map[string]string {
	return hv.Vars
}

// ReplaceVars replace vars in the input string.
func (hv *HelpVars) ReplaceVars(input string) string {
	// if not use var
	if !strings.Contains(input, "{$") {
		return input
	}

	var ss []string
	for n, v := range hv.Vars {
		ss = append(ss, fmt.Sprintf(HelpVarFormat, n), v)
	}

	return strings.NewReplacer(ss...).Replace(input)
}

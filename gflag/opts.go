package gflag

import (
	"flag"
	"fmt"
	"strings"

	"github.com/gookit/gcli/v3/helper"
	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/structs"
	"github.com/gookit/goutil/strutil"
)

const (
	shortSepRune = ','
	shortSepChar = ","
)

// DefaultOptWidth for render help
var DefaultOptWidth = 20

// CliOpts cli options management
type CliOpts struct {
	// name inherited from gcli.Command
	name string

	// the options flag set
	fSet *FlagSet
	// metadata for all options, key is option name.
	opts map[string]*CliOpt // TODO support option category
	// all cli option names, without short names.
	//
	// format: {name: length} // TODO delete, move len to opts.
	names map[string]int
	// short names map for options. format: {short: name}
	//
	// eg. {"n": "name", "o": "opt"}
	shorts map[string]string
	// support option category
	categories []OptCategory
	// flag name max length. useful for render help
	// eg: "-V, --version" length is 13
	optMaxLen int
	// exist short names. useful for render help
	hasShort bool
}

// InitFlagSet create and init flag.FlagSet
func (co *CliOpts) InitFlagSet() {
	if co.fSet != nil {
		return
	}

	if co.name == "" {
		co.name = "flags"
	}
	co.fSet = NewFlagSet(co.name, flag.ContinueOnError)
	// disable output internal error message on parse flags
	// ops.fSet.SetOutput(io.Discard)
	// nothing to do ... render usage on after parsed
	co.fSet.Usage = func() {}
	co.optMaxLen = DefaultOptWidth
}

// SetName for CliArgs
func (co *CliOpts) SetName(name string) {
	co.name = name
}

// FSet get the raw *flag.FlagSet
func (co *CliOpts) FSet() *FlagSet { return co.fSet }

// SetFlagSet set the raw *FlagSet
func (co *CliOpts) SetFlagSet(fSet *FlagSet) { co.fSet = fSet }

/***********************************************************************
 * Options:
 * - binding option var
 ***********************************************************************/

// --- bool option

// Bool binding a bool option flag, return pointer
func (co *CliOpts) Bool(name, shorts string, defVal bool, desc string) *bool {
	opt := newOpt(name, desc, defVal, shorts)
	name = co.checkFlagInfo(opt)

	// binding option to flag.FlagSet
	ptr := co.fSet.Bool(name, defVal, opt.Desc)
	opt.flag = co.fSet.Lookup(name)

	return ptr
}

// BoolVar binding a bool option flag
func (co *CliOpts) BoolVar(ptr *bool, opt *CliOpt) { co.boolOpt(ptr, opt) }

// BoolOpt binding a bool option
func (co *CliOpts) BoolOpt(ptr *bool, name, shorts string, defVal bool, desc string) {
	co.boolOpt(ptr, newOpt(name, desc, defVal, shorts))
}

// BoolOpt2 binding a bool option, and allow with CliOptFn for config option.
func (co *CliOpts) BoolOpt2(p *bool, nameAndShorts, desc string, setFns ...CliOptFn) {
	co.boolOpt(p, NewOpt(nameAndShorts, desc, false, setFns...))
}

// binding option and shorts
func (co *CliOpts) boolOpt(ptr *bool, opt *CliOpt) {
	defVal := opt.DValue().Bool()
	name := co.checkFlagInfo(opt)

	opt.flag = co.fSet.BoolVar(ptr, name, defVal, opt.Desc)
}

// --- float option

// Float64Var binding an float64 option flag
func (co *CliOpts) Float64Var(ptr *float64, opt *CliOpt) { co.float64Opt(ptr, opt) }

// Float64Opt binding a float64 option
func (co *CliOpts) Float64Opt(p *float64, name, shorts string, defVal float64, desc string) {
	co.float64Opt(p, newOpt(name, desc, defVal, shorts))
}

func (co *CliOpts) float64Opt(p *float64, opt *CliOpt) {
	defVal := opt.DValue().Float64()
	name := co.checkFlagInfo(opt)

	opt.flag = co.fSet.Float64Var(p, name, defVal, opt.Desc)
}

// --- string option

// Str binding an string option flag, return pointer
func (co *CliOpts) Str(name, shorts string, defVal, desc string) *string {
	opt := newOpt(name, desc, defVal, shorts)
	name = co.checkFlagInfo(opt)

	p := co.fSet.String(name, defVal, opt.Desc)
	opt.flag = co.fSet.Lookup(name)

	return p
}

// StrVar binding an string option flag
func (co *CliOpts) StrVar(p *string, opt *CliOpt) { co.strOpt(p, opt) }

// StrOpt binding a string option.
//
// If defValAndDesc only one elem, will as desc message.
func (co *CliOpts) StrOpt(p *string, name, shorts string, defValAndDesc ...string) {
	var defVal, desc string
	if ln := len(defValAndDesc); ln > 0 {
		if ln >= 2 {
			defVal, desc = defValAndDesc[0], defValAndDesc[1]
		} else { // only one as desc
			desc = defValAndDesc[0]
		}
	}

	co.StrOpt2(p, name, desc, func(opt *CliOpt) {
		opt.DefVal = defVal
		opt.Shorts = strutil.Split(shorts, shortSepChar)
	})
}

// StrOpt2 binding a string option, and allow with CliOptFn for config option.
func (co *CliOpts) StrOpt2(p *string, nameAndShorts, desc string, setFns ...CliOptFn) {
	co.strOpt(p, NewOpt(nameAndShorts, desc, "", setFns...))
}

// binding option and shorts
func (co *CliOpts) strOpt(p *string, opt *CliOpt) {
	defVal := opt.DValue().String()
	name := co.checkFlagInfo(opt)

	// use *p as default value
	if defVal == "" && *p != "" {
		defVal = *p
	}

	opt.flag = co.fSet.StringVar(p, name, defVal, opt.Desc)
}

// --- intX option

// Int binding an int option flag, return pointer
func (co *CliOpts) Int(name, shorts string, defVal int, desc string) *int {
	opt := newOpt(name, desc, defVal, shorts)
	name = co.checkFlagInfo(opt)

	ptr := co.fSet.Int(name, defVal, opt.Desc)
	opt.flag = co.fSet.Lookup(name)

	return ptr
}

// IntVar binding an int option flag
func (co *CliOpts) IntVar(p *int, opt *CliOpt) { co.intOpt(p, opt) }

// IntOpt binding an int option
func (co *CliOpts) IntOpt(p *int, name, shorts string, defVal int, desc string) {
	co.intOpt(p, newOpt(name, desc, defVal, shorts))
}

// IntOpt2 binding an int option and with config func.
func (co *CliOpts) IntOpt2(p *int, nameAndShorts, desc string, setFns ...CliOptFn) {
	opt := newOpt(nameAndShorts, desc, 0, "")
	co.intOpt(p, opt.WithOptFns(setFns...))
}

func (co *CliOpts) intOpt(ptr *int, opt *CliOpt) {
	defVal := opt.DValue().Int()
	name := co.checkFlagInfo(opt)

	// use *p as default value
	if defVal == 0 && *ptr != 0 {
		defVal = *ptr
	}

	opt.flag = co.fSet.IntVar(ptr, name, defVal, opt.Desc)
}

// Int64 binding an int64 option flag, return pointer
func (co *CliOpts) Int64(name, shorts string, defVal int64, desc string) *int64 {
	opt := newOpt(name, desc, defVal, shorts)
	name = co.checkFlagInfo(opt)

	p := co.fSet.Int64(name, defVal, opt.Desc)
	opt.flag = co.fSet.Lookup(name)
	return p
}

// Int64Var binding an int64 option flag
func (co *CliOpts) Int64Var(ptr *int64, opt *CliOpt) { co.int64Opt(ptr, opt) }

// Int64Opt binding an int64 option
func (co *CliOpts) Int64Opt(ptr *int64, name, shorts string, defValue int64, desc string) {
	co.int64Opt(ptr, newOpt(name, desc, defValue, shorts))
}

func (co *CliOpts) int64Opt(ptr *int64, opt *CliOpt) {
	defVal := opt.DValue().Int64()
	name := co.checkFlagInfo(opt)

	// use *p as default value
	if defVal == 0 && *ptr != 0 {
		defVal = *ptr
	}

	opt.flag = co.fSet.Int64Var(ptr, name, defVal, opt.Desc)
}

// --- uintX option

// Uint binding an int option flag, return pointer
func (co *CliOpts) Uint(name, shorts string, defVal uint, desc string) *uint {
	opt := newOpt(name, desc, defVal, shorts)
	name = co.checkFlagInfo(opt)

	ptr := co.fSet.Uint(name, defVal, opt.Desc)
	opt.flag = co.fSet.Lookup(name)

	return ptr
}

// UintVar binding an uint option flag
func (co *CliOpts) UintVar(ptr *uint, opt *CliOpt) { co.uintOpt(ptr, opt) }

// UintOpt binding an uint option
func (co *CliOpts) UintOpt(ptr *uint, name, shorts string, defValue uint, desc string) {
	co.uintOpt(ptr, newOpt(name, desc, defValue, shorts))
}

func (co *CliOpts) uintOpt(ptr *uint, opt *CliOpt) {
	defVal := opt.DValue().Int()
	name := co.checkFlagInfo(opt)

	opt.flag = co.fSet.UintVar(ptr, name, uint(defVal), opt.Desc)
}

// Uint64 binding an int option flag, return pointer
func (co *CliOpts) Uint64(name, shorts string, defVal uint64, desc string) *uint64 {
	opt := newOpt(name, desc, defVal, shorts)
	name = co.checkFlagInfo(opt)

	ptr := co.fSet.Uint64(name, defVal, opt.Desc)
	opt.flag = co.fSet.Lookup(name)

	return ptr
}

// Uint64Var binding an uint option flag
func (co *CliOpts) Uint64Var(ptr *uint64, opt *CliOpt) { co.uint64Opt(ptr, opt) }

// Uint64Opt binding an uint64 option
func (co *CliOpts) Uint64Opt(ptr *uint64, name, shorts string, defVal uint64, desc string) {
	co.uint64Opt(ptr, newOpt(name, desc, defVal, shorts))
}

// binding option and shorts
func (co *CliOpts) uint64Opt(ptr *uint64, opt *CliOpt) {
	defVal := opt.DValue().Int64()
	name := co.checkFlagInfo(opt)

	opt.flag = co.fSet.Uint64Var(ptr, name, uint64(defVal), opt.Desc)
}

// FuncOptFn func option flag func type
type FuncOptFn func(string) error

// FuncOpt binding a func option flag
//
// Usage:
//
//	cmd.FuncOpt("name", "description ...", func(s string) error {
//		// do something ...
//		return nil
//	})
func (co *CliOpts) FuncOpt(nameAndShorts, desc string, fn FuncOptFn, setFns ...CliOptFn) {
	opt := NewOpt(nameAndShorts, desc, nil, setFns...)
	name := co.checkFlagInfo(opt)

	opt.flag = co.fSet.Func(name, opt.Desc, fn)
}

// Var binding an custom var option flag
func (co *CliOpts) Var(ptr flag.Value, opt *CliOpt) { co.varOpt(ptr, opt) }

// VarOpt binding a custom var option
//
// Usage:
//
//	var names gcli.Strings
//	cmd.VarOpt(&names, "tables", "t", "description ...")
func (co *CliOpts) VarOpt(v flag.Value, name, shorts, desc string) {
	co.varOpt(v, newOpt(name, desc, nil, shorts))
}

// VarOpt2 binding an int option and with config func.
func (co *CliOpts) VarOpt2(v flag.Value, nameAndShorts, desc string, setFns ...CliOptFn) {
	co.varOpt(v, NewOpt(nameAndShorts, desc, nil, setFns...))
}

// binding option and shorts
func (co *CliOpts) varOpt(v flag.Value, opt *CliOpt) {
	name := co.checkFlagInfo(opt)

	opt.flag = co.fSet.Var(v, name, opt.Desc)
}

// check flag option name and short-names
func (co *CliOpts) checkFlagInfo(opt *CliOpt) string {
	// check flag name
	name := opt.initCheck()
	if _, ok := co.opts[name]; ok {
		helper.Panicf("redefined option flag '%s'", name)
	}

	// NOTICE: must init some required fields
	if co.names == nil {
		co.names = map[string]int{}
		co.opts = map[string]*CliOpt{}
		co.InitFlagSet()
	}

	// is a short name
	helpLen := opt.helpNameLen()
	// fix: must exclude Hidden option
	if !opt.Hidden {
		// +6: type placeholder width
		co.optMaxLen = mathutil.MaxInt(co.optMaxLen, helpLen+6)
	}

	// check short names
	co.checkShortNames(name, opt.Shorts)

	// update name length
	co.names[name] = helpLen
	// storage opt and name
	co.opts[name] = opt
	return name
}

// check short names
func (co *CliOpts) checkShortNames(name string, shorts []string) {
	if len(shorts) == 0 {
		return
	}

	co.hasShort = true
	if co.shorts == nil {
		co.shorts = map[string]string{}
	}

	for _, short := range shorts {
		if name == short {
			helper.Panicf("short name '%s' has been used as the current option name", short)
		}

		if _, ok := co.names[short]; ok {
			helper.Panicf("short name '%s' has been used as an option name", short)
		}

		if n, ok := co.shorts[short]; ok {
			helper.Panicf("short name '%s' has been used by option '%s'", short, n)
		}

		// storage short name
		co.shorts[short] = name
	}
}

/***********************************************************************
 * Options parse
 ***********************************************************************/

// ParseOpts parse options from input args
func (co *CliOpts) ParseOpts(args []string) (err error) {
	// parse options
	if err = co.fSet.Parse(args); err != nil {
		return
	}

	// call options validations
	for _, opt := range co.opts {
		err = opt.Validate(opt.flag.Value.String())
		if err != nil {
			return err
		}
	}
	return
}

/***********************************************************************
 * Options:
 * - helper methods
 ***********************************************************************/

// IterAll Iteration all flag options with metadata
func (co *CliOpts) IterAll(fn func(f *flag.Flag, opt *CliOpt)) {
	co.fSet.VisitAll(func(f *flag.Flag) {
		if _, ok := co.opts[f.Name]; ok {
			fn(f, co.opts[f.Name])
		}
	})
}

// ShortNames get all short-names of the option
func (co *CliOpts) ShortNames(name string) (ss []string) {
	if opt, ok := co.opts[name]; ok {
		ss = opt.Shorts
	}
	return
}

// IsShortOpt alias of the IsShortcut()
func (co *CliOpts) IsShortOpt(short string) bool { return co.IsShortName(short) }

// IsShortName check it is a shortcut name
func (co *CliOpts) IsShortName(short string) bool {
	if len(short) != 1 {
		return false
	}

	_, ok := co.shorts[short]
	return ok
}

// IsOption check it is an option name
func (co *CliOpts) IsOption(name string) bool { return co.HasOption(name) }

// HasOption check it is an option name
func (co *CliOpts) HasOption(name string) bool {
	_, ok := co.names[name]
	return ok
}

// LookupFlag get flag.Flag by name
func (co *CliOpts) LookupFlag(name string) *flag.Flag { return co.fSet.Lookup(name) }

// Opt get CliOpt by name
func (co *CliOpts) Opt(name string) *CliOpt { return co.opts[name] }

// Opts get all flag options
func (co *CliOpts) Opts() map[string]*CliOpt { return co.opts }

/***********************************************************************
 * flag options metadata
 ***********************************************************************/

// CliOptFn opt config func type
type CliOptFn func(opt *CliOpt)

// WithRequired setting for option
func WithRequired() CliOptFn {
	return func(opt *CliOpt) { opt.Required = true }
}

// WithDefault value setting for option
func WithDefault(defVal any) CliOptFn {
	return func(opt *CliOpt) { opt.DefVal = defVal }
}

// WithShorts setting for option
func WithShorts(shorts ...string) CliOptFn {
	return func(opt *CliOpt) { opt.Shorts = shorts }
}

// WithShortcut setting for option
func WithShortcut(shortcut string) CliOptFn {
	return func(opt *CliOpt) { opt.Shorts = strutil.Split(shortcut, shortSepChar) }
}

// WithValidator setting for option
func WithValidator(fn func(val string) error) CliOptFn {
	return func(opt *CliOpt) { opt.Validator = fn }
}

// CliOpt define for a flag option
type CliOpt struct {
	// go flag value
	flag *flag.Flag
	// Name of flag and description
	Name, Desc string
	// default value for the flag option
	DefVal any
	// wrapped the default value
	defVal *structs.Value
	// Shorts shorthand names. eg: ["o", "a"]
	Shorts []string
	// EnvVar allow set flag value from ENV var
	EnvVar string

	// --- advanced settings

	// Hidden the option on help
	Hidden bool
	// Required the option is required
	Required bool
	// Validator support validate the option flag value
	Validator func(val string) error
	// TODO interactive question for collect value
	Question string
}

// NewOpt quick create an CliOpt instance
func NewOpt(nameAndShorts, desc string, defVal any, setFns ...CliOptFn) *CliOpt {
	return newOpt(nameAndShorts, desc, defVal, "").WithOptFns(setFns...)
}

// newOpt quick create an CliOpt instance
func newOpt(nameAndShorts, desc string, defVal any, shortcut string) *CliOpt {
	return &CliOpt{
		Name: nameAndShorts,
		Desc: desc,
		// other info
		DefVal: defVal,
		Shorts: strutil.Split(shortcut, shortSepChar),
	}
}

// WithOptFns set for current option
func (m *CliOpt) WithOptFns(fns ...CliOptFn) *CliOpt {
	for _, fn := range fns {
		fn(m)
	}
	return m
}

func (m *CliOpt) initCheck() string {
	// feat: support add shorts by option name. eg: "name,n"
	if strings.ContainsRune(m.Name, shortSepRune) {
		ss := strings.Split(m.Name, shortSepChar)
		m.Name = ss[0]
		m.Shorts = append(m.Shorts, ss[1:]...)
	}

	if m.Desc != "" {
		desc := strings.Trim(m.Desc, "; ")
		if strings.ContainsRune(desc, ';') {
			// format: desc;required
			// format: desc;required;env TODO parse ENV var
			parts := strutil.SplitNTrimmed(desc, ";", 2)
			if ln := len(parts); ln > 1 {
				bl, err := strutil.Bool(parts[1])
				if err == nil && bl {
					desc = parts[0]
					m.Required = true
				}
			}
		}

		m.Desc = desc
	}

	// filter shorts
	if len(m.Shorts) > 0 {
		m.Shorts = cflag.FilterNames(m.Shorts)
	}
	return m.goodName()
}

// good name of the flag
func (m *CliOpt) goodName() string {
	name := strings.Trim(m.Name, "- ")
	if name == "" {
		helper.Panicf("option flag name cannot be empty")
	}

	if !helper.IsGoodName(name) {
		helper.Panicf("option flag name '%s' is invalid, must match: %s", name, helper.RegGoodName)
	}

	// update self name
	m.Name = name
	return name
}

// Shorts2String join shorts to a string
func (m *CliOpt) Shorts2String(sep ...string) string { return m.ShortsString(sep...) }

// ShortsString join shorts to a string
func (m *CliOpt) ShortsString(sep ...string) string {
	if len(m.Shorts) == 0 {
		return ""
	}
	return strings.Join(m.Shorts, sepStr(sep))
}

// HelpName for show help
func (m *CliOpt) HelpName() string {
	return cflag.AddPrefixes(m.Name, m.Shorts)
}

func (m *CliOpt) helpNameLen() int {
	return len(m.HelpName())
}

// Validate the binding value
func (m *CliOpt) Validate(val string) error {
	if m.Required && val == "" {
		return fmt.Errorf("option '%s' is required", m.Name)
	}

	// call user custom validator
	if m.Validator != nil {
		return m.Validator(val)
	}
	return nil
}

// Flag value
func (m *CliOpt) Flag() *flag.Flag {
	return m.flag
}

// DValue wrap the default value
func (m *CliOpt) DValue() *structs.Value {
	if m.defVal == nil {
		m.defVal = &structs.Value{V: m.DefVal}
	}
	return m.defVal
}

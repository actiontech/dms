// This package implements a parser for configuration files.
// This allows easy reading and writing of structured configuration files.
//
// Given the configuration file:
//
//	[default]
//	host = example.com
//	port = 443
//	php = on
//
//	[service-1]
//	host = s1.example.com
//	allow-writing = false
//
// To read this configuration file, do:
//
//	c, err := config.ReadConfigFile("server.conf")
//	c.GetString("default", "host")               // returns example.com
//	c.GetInt("", "port")                         // returns 443 (assumes "default")
//	c.GetBool("", "php")                         // returns true
//	c.GetString("service-1", "host")             // returns s1.example.com
//	c.GetBool("service-1","allow-writing")       // returns false
//	c.GetInt("service-1", "port")                // returns 0 and a GetError
//
// Note that all section and option names are case insensitive. All values are case
// sensitive.
//
// Goconfig's string substitution syntax has not been removed. However, it may be
// taken out or modified in the future.
package config

import (
	"fmt"
	"strings"
)

// ConfigFile is the representation of configuration settings.
// The public interface is entirely through methods.
type ConfigFile struct {
	data         map[string][]*KVData // Maps sections to options to values.
	sectionOrder []string             // comments and option order
}

type KVData struct {
	key   string
	value string
}

const (
	// Get Errors
	SectionNotFound = iota
	OptionNotFound
	MaxDepthReached

	// Read Errors
	BlankSection

	// Get and Read Errors
	CouldNotParse
)

const (
	//line category
	emptyLine   = "emptyline"
	commentLine = "commentline"
)
const virtualSection = "virtualsection" //to save comments before first real section eg: comments

var (
	DefaultSection = "default" // Default section name (must be lower-case).
	DepthValues    = 200       // Maximum allowed depth when recursively substituing variable names.

	// Strings accepted as bool.
	BoolStrings = map[string]bool{
		"t":     true,
		"true":  true,
		"y":     true,
		"yes":   true,
		"on":    true,
		"1":     true,
		"f":     false,
		"false": false,
		"n":     false,
		"no":    false,
		"off":   false,
		"0":     false,
		"":      true, //Only one key is considered as true
	}
)

// AddSection adds a new section to the configuration.
// It returns true if the new section was inserted, and false if the section already existed.
func (c *ConfigFile) AddSection(section string) bool {
	section = strings.ToLower(section)

	if _, ok := c.data[section]; ok {
		return false
	}

	c.data[section] = []*KVData{}
	c.sectionOrder = append(c.sectionOrder, section)
	return true
}

// RemoveSection removes a section from the configuration.
// It returns true if the section was removed, and false if section did not exist.
func (c *ConfigFile) RemoveSection(section string) bool {
	section = strings.ToLower(section)

	switch _, ok := c.data[section]; {
	case !ok:
		return false
	case section == DefaultSection:
		return false // default section cannot be removed
	default:
		delete(c.data, section)
	}

	return true
}

// AddOption adds a new option and value to the configuration, replace value if key exists by default.
// It returns true if the option and value were inserted, and false if the value was overwritten.
// If the section does not exist in advance, it is created.
// If the section has duplicated options, all options would be overwritten
func (c *ConfigFile) AddOption(section string, option string, value string, replaceArg ...bool) bool {
	c.AddSection(section) // make sure section exists

	section = strings.ToLower(section)
	option = strings.ToLower(option)
	replace := true
	if len(replaceArg) != 0 {
		replace = replaceArg[0]
	}

	if replace {
		isOptionExisted := false
		for i, kvData := range c.data[section] {
			if kvData.key == option && option != emptyLine &&
				option != commentLine {
				isOptionExisted = true
				c.data[section][i] = &KVData{
					key:   option,
					value: value,
				}
			}
		}

		if !isOptionExisted {
			c.data[section] = append(c.data[section], &KVData{
				key:   option,
				value: value,
			})
		}
		return !isOptionExisted
	}
	c.data[section] = append(c.data[section], &KVData{
		key:   option,
		value: value,
	})
	return true
}

// RemoveOption removes a option and value from the configuration.
// It returns true if the option and value were removed, and false otherwise,
// including if the section did not exist.
func (c *ConfigFile) RemoveOption(section string, option string) bool {
	section = strings.ToLower(section)
	option = strings.ToLower(option)

	if _, ok := c.data[section]; !ok {
		return false
	}

	kvDatas := c.data[section]
	for i := 0; i < len(kvDatas); i++ {
		if kvDatas[i].key == option {
			kvDatas = append(kvDatas[:i], kvDatas[i+1:]...)
			i--
		}
	}
	c.data[section] = kvDatas

	return true
}

// NewConfigFile creates an empty configuration representation.
// This representation can be filled with AddSection and AddOption and then
// saved to a file using WriteConfigFile.
func NewConfigFile() *ConfigFile {
	c := new(ConfigFile)
	c.data = make(map[string][]*KVData)
	c.AddSection(DefaultSection) // default section always exists
	return c
}

type GetError struct {
	Reason    int
	ValueType string
	Value     string
	Section   string
	Option    string
}

func (err GetError) Error() string {
	switch err.Reason {
	case SectionNotFound:
		return fmt.Sprintf("section '%s' not found", err.Section)
	case OptionNotFound:
		return fmt.Sprintf("option '%s' not found in section '%s'", err.Option, err.Section)
	case CouldNotParse:
		return fmt.Sprintf("could not parse %s value '%s'", err.ValueType, err.Value)
	case MaxDepthReached:
		return fmt.Sprintf("possible cycle while unfolding variables: max depth of %d reached", DepthValues)
	}

	return "invalid get error"
}

type ReadError struct {
	Reason int
	Line   string
}

func (err ReadError) Error() string {
	switch err.Reason {
	case BlankSection:
		return "empty section name not allowed"
	case CouldNotParse:
		return fmt.Sprintf("could not parse line: %s", err.Line)
	}

	return "invalid read error"
}
